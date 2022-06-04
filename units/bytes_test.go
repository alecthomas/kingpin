package units

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase2BytesString(t *testing.T) {
	assert.Equal(t, Base2Bytes(0).String(), "0B")
	assert.Equal(t, Base2Bytes(1025).String(), "1KiB1B")
	assert.Equal(t, Base2Bytes(1048577).String(), "1MiB1B")
}

func TestParseBase2Bytes(t *testing.T) {
	n, err := ParseBase2Bytes("0B")
	assert.NoError(t, err)
	assert.Equal(t, 0, int(n))
	_, err = ParseBase2Bytes("1kB")
	assert.Error(t, err)
	n, err = ParseBase2Bytes("1KB")
	assert.NoError(t, err)
	assert.Equal(t, 1024, int(n))
	n, err = ParseBase2Bytes("1MB1KB25B")
	assert.NoError(t, err)
	assert.Equal(t, 1049625, int(n))
	n, err = ParseBase2Bytes("1.5MB")
	assert.NoError(t, err)
	assert.Equal(t, 1572864, int(n))

	_, err = ParseBase2Bytes("1kiB")
	assert.Error(t, err)
	n, err = ParseBase2Bytes("1KiB")
	assert.NoError(t, err)
	assert.Equal(t, 1024, int(n))
	n, err = ParseBase2Bytes("1MiB1KiB25B")
	assert.NoError(t, err)
	assert.Equal(t, 1049625, int(n))
	n, err = ParseBase2Bytes("1.5MiB")
	assert.NoError(t, err)
	assert.Equal(t, 1572864, int(n))
}

func TestBase2BytesUnmarshalText(t *testing.T) {
	var n Base2Bytes
	err := n.UnmarshalText([]byte("0B"))
	assert.NoError(t, err)
	assert.Equal(t, 0, int(n))
	err = n.UnmarshalText([]byte("1kB"))
	assert.Error(t, err)
	err = n.UnmarshalText([]byte("1KB"))
	assert.NoError(t, err)
	assert.Equal(t, 1024, int(n))
	err = n.UnmarshalText([]byte("1MB1KB25B"))
	assert.NoError(t, err)
	assert.Equal(t, 1049625, int(n))
	err = n.UnmarshalText([]byte("1.5MB"))
	assert.NoError(t, err)
	assert.Equal(t, 1572864, int(n))

	err = n.UnmarshalText([]byte("1kiB"))
	assert.Error(t, err)
	err = n.UnmarshalText([]byte("1KiB"))
	assert.NoError(t, err)
	assert.Equal(t, 1024, int(n))
	err = n.UnmarshalText([]byte("1MiB1KiB25B"))
	assert.NoError(t, err)
	assert.Equal(t, 1049625, int(n))
	err = n.UnmarshalText([]byte("1.5MiB"))
	assert.NoError(t, err)
	assert.Equal(t, 1572864, int(n))
}

func TestBase2Floor(t *testing.T) {
	var n Base2Bytes = KiB
	assert.Equal(t, "1KiB", n.Floor().String())
	n = MiB + KiB
	assert.Equal(t, "1MiB", n.Floor().String())
	n = GiB + MiB + KiB
	assert.Equal(t, "1GiB", n.Floor().String())
	n = 3*GiB + 2*MiB + KiB
	assert.Equal(t, "3GiB", n.Floor().String())
}

func TestBase2Round(t *testing.T) {
	var n Base2Bytes = KiB
	assert.Equal(t, "1KiB", n.Round(1).String())
	n = MiB + KiB
	assert.Equal(t, "1MiB", n.Round(1).String())
	n = GiB + MiB + KiB
	assert.Equal(t, "1GiB", n.Round(1).String())
	n = 3*GiB + 2*MiB + KiB
	assert.Equal(t, "3GiB", n.Round(1).String())
	n = KiB
	assert.Equal(t, "1KiB", n.Round(2).String())
	n = MiB + KiB
	assert.Equal(t, "1MiB1KiB", n.Round(2).String())
	n = GiB + MiB + KiB
	assert.Equal(t, "1GiB1MiB", n.Round(2).String())
	n = 3*GiB + 2*MiB + KiB
	assert.Equal(t, "3GiB2MiB", n.Round(2).String())
}

func TestMetricBytesString(t *testing.T) {
	assert.Equal(t, MetricBytes(0).String(), "0B")
	// TODO: SI standard prefix is lowercase "kB"
	assert.Equal(t, MetricBytes(1001).String(), "1KB1B")
	assert.Equal(t, MetricBytes(1001025).String(), "1MB1KB25B")
}

func TestParseMetricBytes(t *testing.T) {
	n, err := ParseMetricBytes("0B")
	assert.NoError(t, err)
	assert.Equal(t, 0, int(n))
	n, err = ParseMetricBytes("1kB")
	assert.NoError(t, err)
	assert.Equal(t, 1000, int(n))
	n, err = ParseMetricBytes("1KB1B")
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(n))
	n, err = ParseMetricBytes("1MB1KB25B")
	assert.NoError(t, err)
	assert.Equal(t, 1001025, int(n))
	n, err = ParseMetricBytes("1.5MB")
	assert.NoError(t, err)
	assert.Equal(t, 1500000, int(n))
}

func TestParseStrictBytes(t *testing.T) {
	n, err := ParseStrictBytes("0B")
	assert.NoError(t, err)
	assert.Equal(t, 0, int(n))

	_, err = ParseStrictBytes("1kiB")
	assert.Error(t, err)
	n, err = ParseStrictBytes("1KiB")
	assert.NoError(t, err)
	assert.Equal(t, 1024, int(n))
	n, err = ParseStrictBytes("1MiB1KiB25B")
	assert.NoError(t, err)
	assert.Equal(t, 1049625, int(n))
	n, err = ParseStrictBytes("1.5MiB")
	assert.NoError(t, err)
	assert.Equal(t, 1572864, int(n))

	n, err = ParseStrictBytes("0B")
	assert.NoError(t, err)
	assert.Equal(t, 0, int(n))
	n, err = ParseStrictBytes("1kB")
	assert.NoError(t, err)
	assert.Equal(t, 1000, int(n))
	n, err = ParseStrictBytes("1KB1B")
	assert.NoError(t, err)
	assert.Equal(t, 1001, int(n))
	n, err = ParseStrictBytes("1MB1KB25B")
	assert.NoError(t, err)
	assert.Equal(t, 1001025, int(n))
	n, err = ParseStrictBytes("1.5MB")
	assert.NoError(t, err)
	assert.Equal(t, 1500000, int(n))
}

func TestMetricFloor(t *testing.T) {
	var n MetricBytes = KB
	assert.Equal(t, "1KB", n.Floor().String())
	n = MB + KB
	assert.Equal(t, "1MB", n.Floor().String())
	n = GB + MB + KB
	assert.Equal(t, "1GB", n.Floor().String())
	n = 3*GB + 2*MB + KB
	assert.Equal(t, "3GB", n.Floor().String())
}

func TestMetricRound(t *testing.T) {
	var n MetricBytes = KB
	assert.Equal(t, "1KB", n.Round(1).String())
	n = MB + KB
	assert.Equal(t, "1MB", n.Round(1).String())
	n = GB + MB + KB
	assert.Equal(t, "1GB", n.Round(1).String())
	n = 3*GB + 2*MB + KB
	assert.Equal(t, "3GB", n.Round(1).String())
	n = KB
	assert.Equal(t, "1KB", n.Round(2).String())
	n = MB + KB
	assert.Equal(t, "1MB1KB", n.Round(2).String())
	n = GB + MB + KB
	assert.Equal(t, "1GB1MB", n.Round(2).String())
	n = 3*GB + 2*MB + KB
	assert.Equal(t, "3GB2MB", n.Round(2).String())
}

func TestJSON(t *testing.T) {
	type args struct {
		b Base2Bytes
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "0B",
			args: args{
				b: 0,
			},
		},
		{
			name: "1B",
			args: args{
				b: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.args.b)
			assert.NoError(t, err)
			var b Base2Bytes
			err = json.Unmarshal(data, &b)
			assert.NoError(t, err)
			assert.Equal(t, tt.args.b.String(), b.String())
		})
	}
}
