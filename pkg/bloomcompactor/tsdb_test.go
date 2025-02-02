package bloomcompactor

import (
	"context"
	"math"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	v1 "github.com/grafana/loki/pkg/storage/bloom/v1"
	"github.com/grafana/loki/pkg/storage/stores/shipper/indexshipper/tsdb/index"
)

type forSeriesTestImpl []*v1.Series

func (f forSeriesTestImpl) ForSeries(
	_ context.Context,
	_ index.FingerprintFilter,
	_ model.Time,
	_ model.Time,
	fn func(labels.Labels, model.Fingerprint, []index.ChunkMeta),
	_ ...*labels.Matcher,
) error {
	for i := range f {
		unmapped := make([]index.ChunkMeta, 0, len(f[i].Chunks))
		for _, c := range f[i].Chunks {
			unmapped = append(unmapped, index.ChunkMeta{
				MinTime:  int64(c.Start),
				MaxTime:  int64(c.End),
				Checksum: c.Checksum,
			})
		}

		fn(nil, f[i].Fingerprint, unmapped)
	}
	return nil
}

func (f forSeriesTestImpl) Close() error {
	return nil
}

func TestTSDBSeriesIter(t *testing.T) {
	input := []*v1.Series{
		{
			Fingerprint: 1,
			Chunks: []v1.ChunkRef{
				{
					Start:    0,
					End:      1,
					Checksum: 2,
				},
				{
					Start:    3,
					End:      4,
					Checksum: 5,
				},
			},
		},
	}
	srcItr := v1.NewSliceIter(input)
	itr := NewTSDBSeriesIter(context.Background(), forSeriesTestImpl(input), v1.NewBounds(0, math.MaxUint64))

	v1.EqualIterators[*v1.Series](
		t,
		func(a, b *v1.Series) {
			require.Equal(t, a, b)
		},
		itr,
		srcItr,
	)
}

func TestTSDBSeriesIter_Expiry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	itr := NewTSDBSeriesIter(ctx, forSeriesTestImpl{
		{}, // a single entry
	}, v1.NewBounds(0, math.MaxUint64))

	require.False(t, itr.Next())
	require.Error(t, itr.Err())

}
