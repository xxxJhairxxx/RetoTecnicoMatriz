package matrix

import (
	"errors"
	"math"
)

type Matrix [][]float64

var (
	ErrEmpty       = errors.New("la matriz no puede estar vacia")
	ErrRagged      = errors.New("todas las filas deben tener la misma cantidad de columnas")
	ErrNotFinite   = errors.New("la matriz solo admite numeros finitos")
	ErrTooWide     = errors.New("se requiere una matriz con filas >= columnas (m >= n)")
	ErrRankDeficit = errors.New("las columnas de la matriz son linealmente dependientes; no existe factorizacion QR de rango completo")
)

func (a Matrix) Validate() error {
	if len(a) == 0 || len(a[0]) == 0 {
		return ErrEmpty
	}
	n := len(a[0])
	for _, row := range a {
		if len(row) != n {
			return ErrRagged
		}
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return ErrNotFinite
			}
		}
	}
	return nil
}

func (a Matrix) QR() (q, r Matrix, err error) {
	if err = a.Validate(); err != nil {
		return nil, nil, err
	}
	m, n := len(a), len(a[0])
	if m < n {
		return nil, nil, ErrTooWide
	}

	q = make(Matrix, m)
	for i := range q {
		q[i] = make([]float64, n)
		copy(q[i], a[i])
	}
	r = make(Matrix, n)
	for i := range r {
		r[i] = make([]float64, n)
	}

	for j := 0; j < n; j++ {
		for i := 0; i < j; i++ {
			var dot float64
			for k := 0; k < m; k++ {
				dot += q[k][i] * q[k][j]
			}
			r[i][j] = dot
			for k := 0; k < m; k++ {
				q[k][j] -= dot * q[k][i]
			}
		}

		var norm float64
		for k := 0; k < m; k++ {
			norm += q[k][j] * q[k][j]
		}
		norm = math.Sqrt(norm)
		if norm <= 1e-12*columnScale(a, j) || norm == 0 {
			return nil, nil, ErrRankDeficit
		}
		r[j][j] = norm
		for k := 0; k < m; k++ {
			q[k][j] /= norm
		}
	}
	return q, r, nil
}

func columnScale(a Matrix, j int) float64 {
	scale := 1.0
	for i := range a {
		if v := math.Abs(a[i][j]); v > scale {
			scale = v
		}
	}
	return scale
}
