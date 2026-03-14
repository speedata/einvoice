package einvoice

import (
	"bytes"
	"os"
	"testing"
)

// Benchmark suite for measuring performance across all formats and profiles.
// Run with: go test -bench=BenchmarkSuite -benchmem
// Created in response to https://github.com/speedata/einvoice/issues/143

// benchmarkFixtures defines representative test files for each format/profile.
var benchmarkFixtures = []struct {
	name string
	file string
}{
	{"CII/Minimum", "testdata/cii/minimum/zugferd-minimum-rechnung.xml"},
	{"CII/Basic", "testdata/cii/basic/zugferd-basic-einfach.xml"},
	{"CII/EN16931", "testdata/cii/en16931/CII_example1.xml"},
	{"CII/Extended", "testdata/cii/extended/zugferd-extended-warenrechnung.xml"},
	{"CII/XRechnung", "testdata/cii/xrechnung/zugferd-xrechnung-einfach.xml"},
	{"UBL/Invoice", "testdata/ubl/invoice/ubl-tc434-example1.xml"},
	{"UBL/CreditNote", "testdata/ubl/creditnote/ubl-tc434-creditnote1.xml"},
	{"PEPPOL/Invoice", "testdata/peppol/valid/base-example.xml"},
}

func BenchmarkSuiteParse(b *testing.B) {
	for _, bm := range benchmarkFixtures {
		data, err := os.ReadFile(bm.file)
		if err != nil {
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for b.Loop() {
				if _, err := ParseReader(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSuiteValidate(b *testing.B) {
	for _, bm := range benchmarkFixtures {
		inv, err := ParseXMLFile(bm.file)
		if err != nil {
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				inv.Validate()
			}
		})
	}
}

func BenchmarkSuiteCalculate(b *testing.B) {
	for _, bm := range benchmarkFixtures {
		inv, err := ParseXMLFile(bm.file)
		if err != nil {
			continue
		}
		if inv.ProfileLevel() < 3 {
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				inv.UpdateApplicableTradeTax(nil)
				inv.UpdateTotals()
			}
		})
	}
}

func BenchmarkSuiteWrite(b *testing.B) {
	for _, bm := range benchmarkFixtures {
		inv, err := ParseXMLFile(bm.file)
		if err != nil {
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			var buf bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				buf.Reset()
				if err := inv.Write(&buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
