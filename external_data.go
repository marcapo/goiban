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

package goiban

import (
	"database/sql"
	"fmt"
	"log"

	data "github.com/marcapo/goiban-data"
	co "github.com/marcapo/goiban/countries"
	"github.com/tealeg/xlsx"
)

var (
	SELECT_BIC                   = "SELECT bic FROM BANK_DATA WHERE bankcode = ? AND country = ?;"
	SELECT_BIC_STMT              *sql.Stmt
	SELECT_BANK_INFORMATION      = "SELECT bankcode, name, zip, city, bic FROM BANK_DATA WHERE bankcode = ? AND country = ?;"
	SELECT_BANK_INFORMATION_STMT *sql.Stmt
)

func GetBic(iban *Iban, intermediateResult *ValidationResult, repo data.BankDataRepository) *ValidationResult {
	length, ok := COUNTRY_CODE_TO_BANK_CODE_LENGTH[(iban.countryCode)]

	if !ok {
		intermediateResult.Messages = append(intermediateResult.Messages, "Cannot get BIC. No information available.")
		return intermediateResult
	}

	if len(iban.bban) < length {
		intermediateResult.Messages = append(intermediateResult.Messages, "Cannot get BIC for BBAN "+iban.bban)
		return intermediateResult
	}

	bankCode := iban.bban[0:length]
	bankData := GetBankInformationByCountryAndBankCodeFromDb(iban.countryCode, bankCode, repo)

	if bankData == nil {
		intermediateResult.Messages = append(intermediateResult.Messages, "No BIC found for bank code: "+bankCode)
		return intermediateResult
	}

	// issue #17 - Custom Rule for Commerzbank
	//
	// See https://www.eckd-kigst.de/fileadmin/user_upload/eckd/Downloads_KFM/Deutsche_Bundesbank_Uebersicht_der_IBAN_Regeln_Stand_Juni_2013.pdf
	if iban.countryCode == "DE" &&
		len(bankData.Bankcode) > 6 &&
		bankData.Bankcode[3:6] == "400" {
		bankData.Bic = "COBADEFFXXX"
	}

	intermediateResult.BankData = *bankData

	return intermediateResult
}

func prepareSelectBankInformationStatement(db *sql.DB) {
	var err error

	SELECT_BANK_INFORMATION_STMT, err = db.Prepare(SELECT_BANK_INFORMATION)
	if err != nil {
		panic("Couldn't prepare statement: " + SELECT_BANK_INFORMATION)
	}

}

func GetBankInformationByCountryAndBankCodeFromDb(countryCode string, bankCode string, repo data.BankDataRepository) *data.BankInfo {

	// if SELECT_BANK_INFORMATION_STMT == nil {
	// 	prepareSelectBankInformationStatement(db)
	// }

	// var dbBankcode, dbName, dbZip, dbCity, dbBic string

	//bankCode = strings.TrimLeft(bankCode, "0")

	data, err := repo.Find(countryCode, bankCode)
	// err := SELECT_BANK_INFORMATION_STMT.QueryRow(bankCode, countryCode).Scan(&dbBankcode, &dbName, &dbZip, &dbCity, &dbBic)

	if err != nil {
		panic("Failed to load bank info from db.")
	}

	return data
}

func prepareSelectBicStatement(db *sql.DB) {
	var err error
	SELECT_BIC_STMT, err = db.Prepare(SELECT_BIC)
	if err != nil {
		panic("Couldn't prepare statement: " + SELECT_BIC)
	}
}

func ReadFileToEntries(path string, t interface{}, out chan interface{}) {
	cLines := make(chan string)
	switch t := t.(type) {
	default:
		fmt.Println("default:", t)
	case *co.AustriaBankFileEntry:
		go readLines(path, cLines)
		var temp string
		temp = <-cLines
		if temp == "" {
			out <- nil
			return
		}
		var num int
		for l := range cLines {
			num++
			if num < 7 { //skip first six lines
				continue
			}
			if len(l) == 0 {
				out <- nil
				return
			}
			out <- co.AustriaBankStringToEntry(l, COUNTRY_CODE_TO_BANK_CODE_LENGTH)
		}
	case *co.BundesbankFileEntry:
		go readLines(path, cLines)
		for l := range cLines {
			if len(l) == 0 {
				out <- nil
				return
			}
			out <- co.BundesbankStringToEntry(l)
		}
	case *co.BelgiumFileEntry:
		file, err := xlsx.FileToSlice(path)
		if err != nil {
			log.Fatalf("Couldn't read belgium file, %v", err)
		}

		rows := file[0]
		// Skip header
		for _, r := range rows[2:] {
			entries := co.BelgiumRowToEntry(r)
			if len(entries) > 0 {
				out <- entries
			}
		}
	case *co.NetherlandsFileEntry:
		file, err := xlsx.FileToSlice(path)
		if err != nil {
			log.Fatalf("Couldn't read netherlands file, %v", err)
		}

		rows := file[0]
		// Skip header
		for _, r := range rows[2:] {
			out <- co.NetherlandsRowToEntry(r)
		}
	case *co.LuxembourgFileEntry:
		file, err := xlsx.FileToSlice(path)
		if err != nil {
			log.Fatalf("Couldn't read luxembourg file, %v", err)
		}

		rows := file[0]
		// Skip header
		for _, r := range rows[2:] {
			out <- co.LuxembourgRowToEntry(r)
		}
	case *co.SwitzerlandFileEntry:
		file, err := xlsx.FileToSlice(path)
		if err != nil {
			log.Fatalf("Couldn't read switzerland file, %v", err)
		}

		rows := file[0]
		// Skip header
		for _, r := range rows[2:] {
			out <- co.SwitzerlandRowToEntry(r, COUNTRY_CODE_TO_BANK_CODE_LENGTH)
		}
	case *co.LiechtensteinFileEntry:
		file, err := xlsx.FileToSlice(path)
		if err != nil {
			out <- nil
			return
		}
		rows := file[0]
		for _, r := range rows[1:] {
			out <- co.LiechtensteinRowToEntry(r, COUNTRY_CODE_TO_BANK_CODE_LENGTH)
		}
	}
	close(out)
}
