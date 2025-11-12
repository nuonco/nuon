package batch

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchFetch(t *testing.T) {
	testCases := map[string]struct {
		pageSize      int
		maxItems      int
		mockFetchFn   func(ctx context.Context, offset, limit int) ([]int, bool, error)
		expectedItems []int
		expectedError error
	}{
		"Successful fetch with single page": {
			pageSize: 10,
			maxItems: 5,
			mockFetchFn: func(ctx context.Context, offset, limit int) ([]int, bool, error) {
				return []int{1, 2, 3, 4, 5}, false, nil
			},
			expectedItems: []int{1, 2, 3, 4, 5},
		},
		"Successful fetch with multiple pages": {
			pageSize: 3,
			maxItems: 7,
			mockFetchFn: func(ctx context.Context, offset, limit int) ([]int, bool, error) {
				switch offset {
				case 0:
					return []int{1, 2, 3}, true, nil
				case 3:
					return []int{4, 5, 6}, true, nil
				case 6:
					return []int{7, 8}, false, nil
				default:
					return nil, false, errors.New("unexpected offset")
				}
			},
			expectedItems: []int{1, 2, 3, 4, 5, 6, 7},
		},
		"Fetch with max items limit": {
			pageSize: 3,
			maxItems: 4,
			mockFetchFn: func(ctx context.Context, offset, limit int) ([]int, bool, error) {
				switch offset {
				case 0:
					return []int{1, 2, 3}, true, nil
				case 3:
					return []int{4, 5, 6}, true, nil
				default:
					return nil, false, errors.New("unexpected offset")
				}
			},
			expectedItems: []int{1, 2, 3, 4},
		},
		"Error during fetch": {
			pageSize: 3,
			maxItems: 10,
			mockFetchFn: func(ctx context.Context, offset, limit int) ([]int, bool, error) {
				return nil, false, errors.New("fetch error")
			},
			expectedError: errors.New("fetch error"),
		},
		"Empty result set": {
			pageSize: 3,
			maxItems: 10,
			mockFetchFn: func(ctx context.Context, offset, limit int) ([]int, bool, error) {
				return []int{}, false, nil
			},
			expectedItems: []int(nil),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			items, err := BatchFetch(ctx, tc.pageSize, tc.maxItems,
				func(ctx context.Context, offset, limit int) ([]int, bool, error) {
					return tc.mockFetchFn(ctx, offset, limit)
				},
			)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedItems, items)
		})
	}
}
