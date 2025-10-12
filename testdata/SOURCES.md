# Test Fixture Sources

Provenance tracking for all test fixtures in this directory.

**Last Updated**: 2025-10-13
**Total Fixtures**: 62 files

## Sources

| Source | Repository/Package | Version/Commit | License | Directories |
|--------|-------------------|----------------|---------|-------------|
| **EN 16931 Test Suite** | [ConnectingEurope/eInvoicing-EN16931](https://github.com/ConnectingEurope/eInvoicing-EN16931) | `a99371b` (2025-10-09) | EUPL 1.2 | `cii/en16931/` (10), `cii/xrechnung/` (1), `ubl/invoice/` (10), `ubl/creditnote/` (1) |
| **ZUGFeRD 2.3.3** | [ferd-net.de](https://www.ferd-net.de/download-zugferd) (ZF233_EN_01) | May 2024 | FeRD License* | `cii/minimum/` (2), `cii/basicwl/` (2), `cii/basic/` (3), `cii/en16931/` (6), `cii/extended/` (4), `cii/xrechnung/` (3) |
| **horstoeko/zugferd** | [horstoeko/zugferd](https://github.com/horstoeko/zugferd) | Latest | MIT | `cii/basic/` (1), `cii/extended/` (2), `negative/malformed/` (2) |
| **UBL 2.1 OASIS** | [OASIS UBL 2.1](https://docs.oasis-open.org/ubl/os-UBL-2.2/xml/) | UBL 2.1 | OASIS Open | `ubl/invoice/` (1), `ubl/creditnote/` (1) |
| **PEPPOL BIS 3.0** | [OpenPEPPOL/peppol-bis-invoice-3](https://github.com/OpenPEPPOL/peppol-bis-invoice-3) | `78d7f7d` (2025-05-29) | OpenPEPPOL | `peppol/valid/` (11) |

\* FeRD License: Free, royalty-free, irrevocable. License text embedded in each XML file.

## Updating Fixtures

When upstream test suites are updated:

```bash
# 1. Clone/update upstream repositories
git clone https://github.com/ConnectingEurope/eInvoicing-EN16931.git
git clone https://github.com/OpenPEPPOL/peppol-bis-invoice-3.git

# 2. Copy relevant files to organized directories (see table above)

# 3. Update commit hashes
cd eInvoicing-EN16931
git log -1 --format="%H %ci"

# 4. Update this file with new commit hashes and date
```

For ZUGFeRD official examples, download the latest package from [ferd-net.de](https://www.ferd-net.de/download-zugferd).

## License Information

All test fixtures are used as test data for validation purposes and are compatible with the project's BSD-3-Clause license.

### License Summary

- **EUPL 1.2**: Copyleft license. Test data usage for validation does not trigger copyleft provisions. [Full text](testdata/en16931-testsuite/LICENSE.txt)
- **FeRD License**: Free, permissive, royalty-free. Embedded in each ZUGFeRD XML file.
- **MIT**: Permissive open source. [Full text](https://opensource.org/licenses/MIT)
- **OASIS Open**: Royalty-free open standard. [Full text](https://docs.oasis-open.org/ubl/UBL-2.1.html)
- **OpenPEPPOL**: Public test examples developed under EU-CEN agreement.

**Legal Note**: Test fixtures are used as input data for testing parser/writer functionality. This does not create derivative works under copyright law. All sources are properly attributed with repository links and commit hashes.
