package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return in
	}

	cur := in
	for _, stage := range stages {
		cur = withDone(cur, done)
		cur = stage(cur)
	}

	return withDone(cur, done)
}

func withDone(in In, done In) Out {
	if done == nil {
		return in
	}

	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				drain(in)
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					drain(in)
					return
				}
			}
		}
	}()
	return out
}

func drain(in In) {
	for {
		if _, ok := <-in; !ok {
			return
		}
	}
}
