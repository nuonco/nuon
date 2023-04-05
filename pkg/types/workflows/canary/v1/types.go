package canaryv1

import "time"

func (s *Step) Duration() time.Duration {
	if s.Interval.GetEndTime() == nil || s.Interval.GetStartTime() == nil {
		return time.Duration(0)
	}

	return s.Interval.GetEndTime().AsTime().Sub(s.Interval.GetStartTime().AsTime())
}
