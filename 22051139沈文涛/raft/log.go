package main

type Log struct {
	Value     int
	Term      int
	Committed bool
}

func (l *Log) Commit() {
	l.Committed = true
}
