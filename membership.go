package fuzzy

import (
	"math"
)

type Membership interface {
	Value(x float64) float64
	Domain() (min float64, max float64)
}

type ConstantMembership struct {
	y float64
}

func (m *ConstantMembership) Value(x float64) float64 {
	return m.y
}

func (m *ConstantMembership) Domain() (float64, float64) {
	return m.y, m.y
}

func Constant(v float64) *ConstantMembership {
	return &ConstantMembership{v}
}

type MinMembership struct {
	memberships []Membership
}

func (m *MinMembership) Value(x float64) float64 {
	min := math.Inf(1)
	for _, mm := range m.memberships {
		min = math.Min(min, mm.Value(x))
	}

	return min
}

func (m *MinMembership) Domain() (float64, float64) {
	return membershipsDomain(m.memberships)
}

func Min(memberships ...Membership) *MinMembership {
	return &MinMembership{memberships}
}

type MaxMembership struct {
	memberships []Membership
}

func (m *MaxMembership) Value(x float64) float64 {
	max := math.Inf(-1)
	for _, mm := range m.memberships {
		max = math.Max(max, mm.Value(x))
	}

	return max
}

func (m *MaxMembership) Domain() (float64, float64) {
	return membershipsDomain(m.memberships)
}

func Max(memberships ...Membership) *MaxMembership {
	return &MaxMembership{memberships}
}

type LinearMembership struct {
	x1 float64
	x2 float64
}

func (m *LinearMembership) Value(x float64) float64 {
	if m.x1 == m.x2 {
		if x < m.x1 {
			return 0
		}
		return 1
	}

	if x < m.x1 {
		return 0
	}

	if x <= m.x2 {
		return (x - m.x1) / (m.x2 - m.x1)
	}

	return 1
}

func (m *LinearMembership) Domain() (float64, float64) {
	return m.x1, m.x2
}

func Linear(x1, x2 float64) *LinearMembership {
	return &LinearMembership{x1, x2}
}

func Step(x1 float64) *LinearMembership {
	return Linear(x1, x1)
}

type TriangularMembership struct {
	x1 float64
	x2 float64
	x3 float64
}

func (m *TriangularMembership) Value(x float64) float64 {
	if m.x1 < x && x < m.x2 {
		return (x - m.x1) / (m.x2 - m.x1)
	}

	if m.x2 <= x && x <= m.x3 {
		return (m.x3 - x) / (m.x3 - m.x2)
	}

	return 0
}

func (m *TriangularMembership) Domain() (float64, float64) {
	return m.x1, m.x3
}

func Triangular(x1, x2, x3 float64) *TriangularMembership {
	return &TriangularMembership{x1, x2, x3}
}

type InvertedMembership struct {
	membership Membership
}

func (m *InvertedMembership) Value(x float64) float64 {
	return 1 - m.membership.Value(x)
}

func (m *InvertedMembership) Domain() (float64, float64) {
	return m.membership.Domain()
}

func Inverted(m Membership) *InvertedMembership {
	return &InvertedMembership{m}
}

type TrapezoidalMembership struct {
	x1 float64
	x2 float64
	x3 float64
	x4 float64
}

func (m *TrapezoidalMembership) Value(x float64) float64 {
	if m.x1 < x && x < m.x2 {
		if m.x1 == m.x2 {
			return 1.0
		}
		return (x - m.x1) / (m.x2 - m.x1)
	}

	if m.x2 <= x && x <= m.x3 {
		return 1.0
	}

	if m.x3 < x && x < m.x4 {
		if m.x3 == m.x4 {
			return 1.0
		}
		return (m.x4 - x) / (m.x4 - m.x3)
	}

	return 0.0
}

func (m *TrapezoidalMembership) Domain() (float64, float64) {
	return m.x1, m.x4
}

func Trapezoid(x1, x2, x3, x4 float64) *TrapezoidalMembership {
	return &TrapezoidalMembership{x1, x2, x3, x4}
}

func membershipsDomain(memberships []Membership) (float64, float64) {
	min := math.Inf(1)
	max := math.Inf(-1)

	for _, mm := range memberships {
		x1, x2 := mm.Domain()
		min = math.Min(min, x1)
		max = math.Max(max, x2)
	}

	return min, max
}
