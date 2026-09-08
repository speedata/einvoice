package rules

//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/validation-1.3.16/cii/schematron/abstract/EN16931-CII-model.sch --version v1.3.16 --package rules --output en16931.go
//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/OpenPEPPOL/peppol-bis-invoice-3/master/rules/sch/PEPPOL-EN16931-CII.sch --version 3.0.19 --package rules --output peppol.go
//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/itplr-kosit/xrechnung-schematron/release-2.4.0/src/validation/schematron/cii/XRechnung-CII-validation.sch --version 2.4.0 --package rules --output xrechnung_cii.go
//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/validation-1.3.16/cii/schematron/preprocessed/EN16931-CII-validation-preprocessed.sch --version v1.3.16 --package rules --syntax-pattern EN16931-CII-Syntax --varname CIISyntaxRules --output cii_syntax.go
//go:generate go run ../cmd/genrules --source https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/validation-1.3.16/ubl/schematron/preprocessed/EN16931-UBL-validation-preprocessed.sch --version v1.3.16 --package rules --syntax-pattern UBL-syntax --varname UBLSyntaxRules --output ubl_syntax.go
