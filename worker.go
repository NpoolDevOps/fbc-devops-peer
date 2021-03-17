package main

type Worker struct {
}

func NewWorkerPeer(config *BasenodeConfig) *Worker {
	return &Worker{}
}

func (n *Worker) Run() error {
	return nil
}
