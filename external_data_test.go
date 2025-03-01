/*
The MIT License (MIT)

Copyright (c) 2014 Chris Grieger

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package goiban_test

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/marcapo/goiban"
	data "github.com/marcapo/goiban-data"
	"github.com/marcapo/goiban-data-loader/loader"
	co "github.com/marcapo/goiban/countries"
)

var (
	repo = data.NewInMemoryStore()
)

func TestMain(m *testing.M) {
	loader.LoadBundesbankData(loader.DefaultBundesbankPath(), repo)
	loader.LoadBelgiumData(loader.DefaultBelgiumPath(), repo)

	retCode := m.Run()

	os.Exit(retCode)
}

func TestCanReadFromAustriaFile(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/austria.csv", &co.AustriaBankFileEntry{}, ch)

	peek := (<-ch).(*co.AustriaBankFileEntry)
	if peek.Name == "" {
		t.Errorf("Failed to read file.")
	}
}

func TestCannotReadFromNonExistingAustriaFile(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/austria_blablablabla.csv", &co.AustriaBankFileEntry{}, ch)
	result := <-ch
	if result != nil {
		t.Errorf("Failed to read file.")
	}
}
func TestCanReadFromBundesbankFile(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/bundesbank.txt", &co.BundesbankFileEntry{}, ch)

	peek := (<-ch).(*co.BundesbankFileEntry)
	if peek.Name == "" {
		t.Errorf("Failed to read file.")
	}
}

func TestCannotReadFromNonExistingBundesbankFile(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/bundesbank_doesntexist.txt", &co.BundesbankFileEntry{}, ch)
	result := <-ch
	if result != nil {
		t.Errorf("Failed to read file.")
	}
}

func TestCanLoadBankInfoFromDatabase(t *testing.T) {
	bankInfo := goiban.GetBankInformationByCountryAndBankCodeFromDb("DE", "84050000", repo)
	fmt.Println(bankInfo)
	if bankInfo == nil {
		t.Errorf("Cannot load data from repo. Is it empty?")
	}
}

func TestCanLoadBankInfoFromDatabaseLeadingZeros(t *testing.T) {
	bankInfo := goiban.GetBankInformationByCountryAndBankCodeFromDb("BE", "001", repo)
	if bankInfo == nil {
		t.Errorf("Cannot load data from repo. Is it empty?")
	}
}

func TestCanReadFromBelgiumXLSX(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/belgium.xlsx", &co.BelgiumFileEntry{}, ch)

	peek := (<-ch).([]co.BelgiumFileEntry)
	if peek[0].Name != "bpost bank" {
		t.Errorf("Failed to read file.")
	}
}

func TestCanReadFromNetherlandsXLSX(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/netherlands.xlsx", &co.NetherlandsFileEntry{}, ch)

	peek := (<-ch).(co.NetherlandsFileEntry)
	if peek.Name != "ABN AMRO BANK N.V" {
		t.Errorf("Failed to read file.")
	}
}
func TestCanReadFromSwitzerlandFile(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/switzerland.xlsx", &co.SwitzerlandFileEntry{}, ch)

	peek := (<-ch).(co.SwitzerlandFileEntry)
	if peek.Bic != "SNBZCHZZXXX" {
		t.Errorf("Failed to read file.")
	}
}

func TestCanReadFromLiechtensteinXLSX(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/liechtenstein.xlsx", &co.LiechtensteinFileEntry{}, ch)
	peek := (<-ch).(co.LiechtensteinFileEntry)
	if peek.Bic != "BALPLI22" {
		t.Errorf("Failed to read file." + peek.Bic)
	}
}
func TestCannotReadFromNonExistingLiechtensteinFile(t *testing.T) {
	ch := make(chan interface{})
	go goiban.ReadFileToEntries("test/lliechtenstein_blablablabla.xlsx", &co.LiechtensteinFileEntry{}, ch)
	result := <-ch
	if result != nil {
		t.Errorf("Failed to read file.")
	}
}

func TestSpecialRuleForCommerzbankBic(t *testing.T) {
	input := "DE12120400000052065002"
	iban := goiban.ParseToIban(input)
	result := goiban.NewValidationResult(true, "", input)

	result = goiban.GetBic(iban, result, repo)

	if result.BankData.Bic != "COBADEFFXXX" {
		t.Errorf("Expected Bic COBADEFFXXX, was %v", result.BankData.Bic)
	}
}
