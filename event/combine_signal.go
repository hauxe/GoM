package event

// OR function combines all signal from channels into a single channel with Or condition
func OR(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}
	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			case <-OR(append(channels[3:], orDone)...):
			}
		}
	}()
	return orDone
}

// AND function combines all signal from channels into a single channel with And condition
func AND(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}
	andDone := make(chan interface{})
	collector := make(chan interface{}, len(channels))
	go func() {
		defer close(andDone)
		switch len(channels) {
		case 2:
			select {
			case collector <- <-channels[0]:
			case collector <- <-channels[1]:
			}
		default:
			select {
			case collector <- <-channels[0]:
			case collector <- <-channels[1]:
			case collector <- <-channels[2]:
			case collector <- <-AND(channels[3:]...):
			}
		}
	}()
	return andDone
}
