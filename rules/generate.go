package rules

//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/master/cii/schematron/abstract/EN16931-CII-model.sch --version v1.3.14.1 --package rules --output en16931.go
//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/OpenPEPPOL/peppol-bis-invoice-3/master/rules/sch/PEPPOL-EN16931-CII.sch --version 3.0.19 --package rules --output peppol.go
