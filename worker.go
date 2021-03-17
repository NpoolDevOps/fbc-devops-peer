package main

type Worker struct {
}

func NewWorkerPeer(config *PeerConfig) *Worker {
	return &Worker{}
}

func (n *Worker) Run() error {
	return nil
}
