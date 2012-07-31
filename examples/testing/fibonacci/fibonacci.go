/*
The fibonacci package serves as both an example of good skynet development
practice (NOTE: IT IS YET TO BE DETERMINED WHAT GOOD SKYNET DEVELOPMENT
PRACTICE IS) and a skynet stress tester. The service will act as a client
and recursively call itself to solve the Fibonacci recurrence.

By collecting the request and response parameters into two publicly
accessible types, it is easy to make sure your client is giving the service
the kind of data it expects, provided the expected version is used.
*/
package fibonacci

type Request struct {
	// The index of the value in the Fibonnaci sequence.
	// F_0 = 0, F_1 = 1, F_{i+2} = F_{i+1} + F_i
	Index int
}

type Response struct {
	// The index of the value in the Fibonacci sequence.
	Index int
	// The numerical value corresponding to the index.
	Value uint64
}
