package pools

type ring struct {
	cnt, i int
	data   []Resource
}

func (rb *ring) Size() int {
	return rb.cnt
}

func (rb *ring) Empty() bool {
	return rb.cnt == 0
}

func (rb *ring) Peek() Resource {
	return rb.data[rb.i]
}

func (rb *ring) Enqueue(x Resource) {
	if rb.cnt >= len(rb.data) {
		rb.grow(2*rb.cnt + 1)
	}
	rb.data[(rb.i+rb.cnt)%len(rb.data)] = x
	rb.cnt++
}

func (rb *ring) Dequeue() (x Resource) {
	x = rb.Peek()
	rb.cnt, rb.i = rb.cnt-1, (rb.i+1)%len(rb.data)
	return
}

func (rb *ring) grow(newSize int) {
	newData := make([]Resource, newSize)

	n := copy(newData, rb.data[rb.i:])
	copy(newData[n:], rb.data[:rb.cnt-n])

	rb.i = 0
	rb.data = newData
}
