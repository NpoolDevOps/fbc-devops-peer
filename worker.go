package main

type Worker struct {
}

func NewWorkerPeer(subRole string, parentSpec string) *Worker {
	return &Worker{}
}

func (n *Worker) Run() error {
	return nil
}
