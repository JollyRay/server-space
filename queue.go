package main

import (
	"sync"
)

type queueS struct {
	Tockens []int64
	mutex   sync.Mutex
}

func (q *queueS) take(tocken int64) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Tockens = append(q.Tockens, tocken)
}

func (q *queueS) takeFirst(tocken int64) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Tockens = append([]int64{0, tocken}, q.Tockens[1:]...)
}

func (q *queueS) leave(tocken int64) {
	q.mutex.Lock()
	for index, toctocken := range q.Tockens {
		if toctocken == tocken {
			q.Tockens = append(q.Tockens[:index], q.Tockens[index+1:]...)
			break
		}
	}
	q.mutex.Unlock()
}

func (q *queueS) next() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.Tockens) > 1 {
		tocken := q.Tockens[1]
		car := cars.getFreeCar(tocken)
		if car != nil {
			tockens.mutex.Lock()
			linker, ok := tockens.List[tocken]
			if ok {
				linker.Car = car
				tockens.List[tocken] = linker
				replyNineRespons(setNineByte(ACCEPT, tocken), *linker.User.conn)
				q.Tockens = append(q.Tockens[:1], q.Tockens[2:]...)
				go imgRepeater(*linker.User.conn, *car.conn, &car.stop, car.size, tocken)
			}
			tockens.mutex.Unlock()
		}
	}
}

var queue queueS = queueS{Tockens: make([]int64, 1)}
