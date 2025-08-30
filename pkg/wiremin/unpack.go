package wiremin

import "fmt"

func UnpackTimeseries(p Payload) (TSMeta, []TSPoint, error) {
	if p.K != KTimeseries {
		return TSMeta{}, nil, fmt.Errorf("wrong kind: %s", p.K)
	}
	if len(p.M) < 5 {
		return TSMeta{}, nil, fmt.Errorf("bad meta")
	}
	meta := TSMeta{
		Symbol: p.M[0].(string), Currency: p.M[1].(string),
		Exchange: p.M[2].(string), Granularity: p.M[3].(string), Timezone: p.M[4].(string),
	}
	pts := make([]TSPoint, 0, len(p.D))
	for _, r := range p.D {
		pts = append(pts, TSPoint{
			T: int64(r[0].(float64)), O: r[1].(float64), H: r[2].(float64),
			L: r[3].(float64), C: r[4].(float64), V: int64(r[5].(float64)),
		})
	}
	return meta, pts, nil
}
