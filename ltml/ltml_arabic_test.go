package ltml

import "testing"

func TestArabicSamples(t *testing.T) {
	samples := []string{
		"test_033_arabic_program",
		"test_055_arabic_index",
	}

	for _, sample := range samples {
		sample := sample
		t.Run(sample, func(t *testing.T) {
			outputFile := sampleOutputFile(sample, t)
			writeSamplePDF(sample, outputFile, t)
		})
	}
}
