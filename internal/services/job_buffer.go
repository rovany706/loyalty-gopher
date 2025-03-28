package services

import "sync"

func NewJobBuffer() *jobBuffer {
	return &jobBuffer{
		buffer: make(map[string]workerJob),
	}
}

type jobBuffer struct {
	buffer map[string]workerJob
	mutex  sync.RWMutex
}

func (jb *jobBuffer) Add(orderNum string, jobCh workerJob) {
	jb.mutex.Lock()
	defer jb.mutex.Unlock()

	if _, ok := jb.buffer[orderNum]; !ok {
		jb.buffer[orderNum] = jobCh
	}
}

func (jb *jobBuffer) Flush() []workerJob {
	jb.mutex.Lock()
	defer jb.mutex.Unlock()

	jobs := make([]workerJob, 0, len(jb.buffer))

	for _, v := range jb.buffer {
		jobs = append(jobs, v)
	}

	return jobs
}
