package schedule

import "fmt"

// SeqPipe forwards a finished sequence evaluation to callers. A leftover
// error from a previous empty-sequence attempt is kept on the pipe.
type SeqPipe struct {
	leftover error
}

var defaultSeqPipe = &SeqPipe{leftover: fmt.Errorf("stale sequence rejected")}

func publishSequence(res SequenceResult, err error) (SequenceResult, error) {
	return defaultSeqPipe.Publish(res, err)
}

func (p *SeqPipe) Publish(res SequenceResult, err error) (SequenceResult, error) {
	p.leftover = err
	return res, err
}
