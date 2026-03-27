package fact

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrFactorizationCancelled = errors.New("cancelled")
	ErrWriterInteraction      = errors.New("writer interaction")
)

type Factorizer interface {
	Factorize(ctx context.Context, numbers []int, writer io.Writer) error
}

type factorizerImpl struct {
	factorizationWorkers int
	writeWorkers         int
}

func New(opts ...FactorizeOption) (*factorizerImpl, error) {
	goMaxProc := runtime.GOMAXPROCS(-1)

	c := &factorizerImpl{
		factorizationWorkers: goMaxProc,
		writeWorkers:         goMaxProc,
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.isValid(); err != nil {
		return nil, err
	}

	return c, nil
}

type FactorizeOption func(*factorizerImpl)

func WithFactorizationWorkers(workers int) FactorizeOption {
	return func(s *factorizerImpl) {
		s.factorizationWorkers = workers
	}
}

func WithWriteWorkers(workers int) FactorizeOption {
	return func(s *factorizerImpl) {
		s.writeWorkers = workers
	}
}

func (f *factorizerImpl) Factorize(
	ctx context.Context,
	numbers []int,
	writer io.Writer,
) error {
	writersWg := new(sync.WaitGroup)
	writerInput := make(chan []string)
	internalDone := make(chan struct{})

	var (
		fErr error
		once sync.Once
	)

	for i := 0; i < f.writeWorkers; i++ {
		writersWg.Go(func() {
			for {
				select {
				case <-internalDone:
					return
				case <-ctx.Done():
					once.Do(func() {
						fErr = errors.Join(ErrFactorizationCancelled, context.Cause(ctx))
					})
				case v, ok := <-writerInput:
					if !ok {
						return
					}

					_, wErr := fmt.Fprintf(writer, "%s = %s\n", v[0], strings.Join(v[1:], " * "))

					if wErr != nil {
						once.Do(func() {
							fErr = fmt.Errorf("%w: writer error: %w", wErr, ErrWriterInteraction)
							close(internalDone)
						})

						return
					}
				}
			}
		})
	}

	factInput := make(chan int)
	factWg := new(sync.WaitGroup)
	for i := 0; i < f.factorizationWorkers; i++ {
		factWg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					once.Do(func() {
						fErr = errors.Join(ErrFactorizationCancelled, context.Cause(ctx))
					})
					return
				case v, ok := <-factInput:
					if !ok {
						return
					}

					select {
					case writerInput <- f.factorization(v):
					case <-ctx.Done():
						return
					case <-internalDone:
						return
					}
				}
			}
		})
	}

l:
	for _, n := range numbers {
		select {
		case <-internalDone:
			break l
		case factInput <- n:
		case <-ctx.Done():
			once.Do(func() {
				fErr = errors.Join(ErrFactorizationCancelled, context.Cause(ctx))
			})
			break l
		}
	}

	close(factInput)
	factWg.Wait()
	close(writerInput)
	writersWg.Wait()

	return fErr
}

func (f *factorizerImpl) isValid() error {
	if f.writeWorkers < 1 {
		return fmt.Errorf("invalid write workers value: %d", f.writeWorkers)
	}

	if f.factorizationWorkers < 1 {
		return fmt.Errorf("invalid factorization workers value: %d", f.factorizationWorkers)
	}

	return nil
}

func (f *factorizerImpl) factorization(n int) []string {
	result := []string{strconv.Itoa(n)}

	if n < 0 {
		result = append(result, "-1")

		if n == math.MinInt {
			n /= 2
			result = append(result, "2")
		}

		n = -n
	}

	if n == 0 || n == 1 {
		result = append(result, strconv.Itoa(n))
		return result
	}

	border := int(math.Sqrt(float64(n)))
	for divider := 2; divider <= border; divider++ {
		if n == 1 {
			return result
		}

		for n%divider == 0 {
			n /= divider
			result = append(result, strconv.Itoa(divider))
		}
	}

	if n != 1 {
		result = append(result, strconv.Itoa(n))
	}

	return result
}
