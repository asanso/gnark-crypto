// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by goff (v0.2.2) DO NOT EDIT

// Package fp contains field arithmetic operations
package fp

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"testing"
)

func TestELEMENTCorrectnessAgainstBigInt(t *testing.T) {
	modulus, _ := new(big.Int).SetString("6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299", 10)
	cmpEandB := func(e *Element, b *big.Int, name string) {
		var _e big.Int
		if e.FromMont().ToBigInt(&_e).Cmp(b) != 0 {
			t.Fatal(name, "failed")
		}
	}
	var modulusMinusOne, one big.Int
	one.SetUint64(1)

	modulusMinusOne.Sub(modulus, &one)

	var n int
	if testing.Short() {
		n = 20
	} else {
		n = 500
	}

	sAdx := supportAdx

	for i := 0; i < n; i++ {
		if i == n/2 && sAdx {
			supportAdx = false // testing without adx instruction
		}
		// sample 2 random big int
		b1, _ := rand.Int(rand.Reader, modulus)
		b2, _ := rand.Int(rand.Reader, modulus)
		rExp := mrand.Uint64()

		// adding edge cases
		// TODO need more edge cases
		switch i {
		case 0:
			rExp = 0
			b1.SetUint64(0)
		case 1:
			b2.SetUint64(0)
		case 2:
			b1.SetUint64(0)
			b2.SetUint64(0)
		case 3:
			rExp = 0
		case 4:
			rExp = 1
		case 5:
			rExp = ^uint64(0) // max uint
		case 6:
			rExp = 2
			b1.Set(&modulusMinusOne)
		case 7:
			b2.Set(&modulusMinusOne)
		case 8:
			b1.Set(&modulusMinusOne)
			b2.Set(&modulusMinusOne)
		}

		rbExp := new(big.Int).SetUint64(rExp)

		var bMul, bAdd, bSub, bDiv, bNeg, bLsh, bInv, bExp, bExp2, bSquare big.Int

		// e1 = mont(b1), e2 = mont(b2)
		var e1, e2, eMul, eAdd, eSub, eDiv, eNeg, eLsh, eInv, eExp, eSquare, eMulAssign, eSubAssign, eAddAssign Element
		e1.SetBigInt(b1)
		e2.SetBigInt(b2)

		// (e1*e2).FromMont() === b1*b2 mod q ... etc
		eSquare.Square(&e1)
		eMul.Mul(&e1, &e2)
		eMulAssign.Set(&e1)
		eMulAssign.MulAssign(&e2)
		eAdd.Add(&e1, &e2)
		eAddAssign.Set(&e1)
		eAddAssign.AddAssign(&e2)
		eSub.Sub(&e1, &e2)
		eSubAssign.Set(&e1)
		eSubAssign.SubAssign(&e2)
		eDiv.Div(&e1, &e2)
		eNeg.Neg(&e1)
		eInv.Inverse(&e1)
		eExp.Exp(e1, rExp)
		eLsh.Double(&e1)

		// same operations with big int
		bAdd.Add(b1, b2).Mod(&bAdd, modulus)
		bMul.Mul(b1, b2).Mod(&bMul, modulus)
		bSquare.Mul(b1, b1).Mod(&bSquare, modulus)
		bSub.Sub(b1, b2).Mod(&bSub, modulus)
		bDiv.ModInverse(b2, modulus)
		bDiv.Mul(&bDiv, b1).
			Mod(&bDiv, modulus)
		bNeg.Neg(b1).Mod(&bNeg, modulus)

		bInv.ModInverse(b1, modulus)
		bExp.Exp(b1, rbExp, modulus)
		bLsh.Lsh(b1, 1).Mod(&bLsh, modulus)

		cmpEandB(&eSquare, &bSquare, "Square")
		cmpEandB(&eMul, &bMul, "Mul")
		cmpEandB(&eMulAssign, &bMul, "MulAssign")
		cmpEandB(&eAdd, &bAdd, "Add")
		cmpEandB(&eAddAssign, &bAdd, "AddAssign")
		cmpEandB(&eSub, &bSub, "Sub")
		cmpEandB(&eSubAssign, &bSub, "SubAssign")
		cmpEandB(&eDiv, &bDiv, "Div")
		cmpEandB(&eNeg, &bNeg, "Neg")
		cmpEandB(&eInv, &bInv, "Inv")
		cmpEandB(&eExp, &bExp, "Exp")

		cmpEandB(&eLsh, &bLsh, "Lsh")

		// legendre symbol
		if e1.Legendre() != big.Jacobi(b1, modulus) {
			t.Fatal("legendre symbol computation failed")
		}
		if e2.Legendre() != big.Jacobi(b2, modulus) {
			t.Fatal("legendre symbol computation failed")
		}

		// these are slow, killing circle ci
		if n <= 5 {
			// sqrt
			var eSqrt, eExp2 Element
			var bSqrt big.Int
			bSqrt.ModSqrt(b1, modulus)
			eSqrt.Sqrt(&e1)
			cmpEandB(&eSqrt, &bSqrt, "Sqrt")

			bits := b2.Bits()
			exponent := make([]uint64, len(bits))
			for k := 0; k < len(bits); k++ {
				exponent[k] = uint64(bits[k])
			}
			eExp2.Exp(e1, exponent...)
			bExp2.Exp(b1, b2, modulus)
			cmpEandB(&eExp2, &bExp2, "Exp multi words")
		}
	}
	supportAdx = sAdx
}

func TestELEMENTIsRandom(t *testing.T) {
	for i := 0; i < 50; i++ {
		var x, y Element
		x.SetRandom()
		y.SetRandom()
		if x.Equal(&y) {
			t.Fatal("2 random numbers are unlikely to be equal")
		}
	}
}

func TestByteElement(t *testing.T) {

	modulus := ElementModulus()

	// test values
	var bs [3][]byte
	r1, _ := rand.Int(rand.Reader, modulus)
	bs[0] = r1.Bytes() // should be r1 as Element
	r2, _ := rand.Int(rand.Reader, modulus)
	r2.Add(modulus, r2)
	bs[1] = r2.Bytes() // should be r2 as Element
	var tmp big.Int
	tmp.SetUint64(0)
	bs[2] = tmp.Bytes() // should be 0 as Element

	// witness values as Element
	var el [3]Element
	el[0].SetBigInt(r1)
	el[1].SetBigInt(r2)
	el[2].SetUint64(0)

	// check conversions
	for i := 0; i < 3; i++ {
		var z Element
		z.SetBytes(bs[i])
		if !z.Equal(&el[i]) {
			t.Fatal("SetBytes fails")
		}
		// check conversion Element to Bytes
		b := z.Bytes()
		z.SetBytes(b)
		if !z.Equal(&el[i]) {
			t.Fatal("Bytes fails")
		}
	}
}

// -------------------------------------------------------------------------------------------------
// benchmarks
// most benchmarks are rudimentary and should sample a large number of random inputs
// or be run multiple times to ensure it didn't measure the fastest path of the function

var benchResElement Element

func BenchmarkInverseELEMENT(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchResElement.Inverse(&x)
	}

}
func BenchmarkExpELEMENT(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Exp(x, mrand.Uint64())
	}
}

func BenchmarkDoubleELEMENT(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Double(&benchResElement)
	}
}

func BenchmarkAddELEMENT(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Add(&x, &benchResElement)
	}
}

func BenchmarkSubELEMENT(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Sub(&x, &benchResElement)
	}
}

func BenchmarkNegELEMENT(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Neg(&benchResElement)
	}
}

func BenchmarkDivELEMENT(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Div(&x, &benchResElement)
	}
}

func BenchmarkFromMontELEMENT(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.FromMont()
	}
}

func BenchmarkToMontELEMENT(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.ToMont()
	}
}
func BenchmarkSquareELEMENT(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Square(&benchResElement)
	}
}

func BenchmarkSqrtELEMENT(b *testing.B) {
	var a Element
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Sqrt(&a)
	}
}

func BenchmarkMulAssignELEMENT(b *testing.B) {
	x := Element{
		14305184132582319705,
		8868935336694416555,
		9196887162930508889,
		15486798265448570248,
		5402985275949444416,
		10893197322525159598,
		3204916688966998390,
		12417238192559061753,
		12426306557607898622,
		1305582522441154384,
		10311846026977660324,
		48736111365249031,
	}
	benchResElement.SetOne()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.MulAssign(&x)
	}
}
