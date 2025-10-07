# Rule Auto-Generation - Current Status

**Branch**: `feature/autogen-rules-en16931`
**Issue**: #27
**Status**: 📋 Planning Complete - Ready to Start Implementation

## Quick Links

- **Full Plan**: See `GENRULES_PLAN.md` for comprehensive implementation details
- **TODO List**: 23 detailed tasks tracked in development session
- **Issue**: https://github.com/speedata/einvoice/issues/27
- **Research**: https://github.com/speedata/einvoice/issues/27#issuecomment-3376614135

## Current Phase: Phase 0 - Planning ✅

- ✅ Research completed
- ✅ Source files identified and validated
- ✅ Implementation plan documented
- ✅ TODO list created (23 tasks)
- ✅ Branch created and pushed

## Next Phase: Phase 1 - Prototype (Tasks 1-8)

### Immediate Next Steps:
1. Create `cmd/genrules/main.go` skeleton
2. Define XML parsing structs
3. Implement basic schematron parser
4. Implement field extraction regex
5. Test with 5-10 sample rules
6. Create `rules/` package directory
7. Generate into separate package

### Expected Output (rules/en16931.go):
```go
package rules

type Rule struct {
    Code        string
    Fields      []string
    Description string
}

var (
    BR1 = Rule{
        Code:        "BR-01",
        Fields:      []string{"BT-24"},
        Description: "An Invoice shall have a Specification identifier (BT-24).",
    }
    // ... 5-10 more sample rules
)
```

### Package Structure:
```
einvoice/
├── rules/              # NEW: Generated rules package
│   └── en16931.go      # Generated EN 16931 rules
├── cmd/genrules/       # Rule generation tool
└── validation.go       # Imports rules package
```

## Source Data

**Official Repository**: https://github.com/ConnectingEurope/eInvoicing-EN16931

**Primary File**:
```
https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/master/cii/schematron/abstract/EN16931-CII-model.sch
```

**Downloaded Test Files** (available locally):
- `/tmp/EN16931-CII-model-abstract.sch` (338 lines, 203 rules)
- `/tmp/EN16931-CII-model.sch` (concrete bindings)

## Key Implementation Details

### XML Structure to Parse:
```xml
<assert test="$BR-XX" flag="fatal" id="BR-XX">
  [BR-XX]-Description text (BT-YY) (BG-ZZ)
</assert>
```

### Extraction Pipeline:
```
Schematron XML → Parse → Extract → Transform → Generate Go Code
     ↓              ↓         ↓          ↓            ↓
   *.sch         xml.Unmarshal  Regex   Clean    template.Execute
```

### Critical Algorithms:

**1. Field Extraction Regex**: `\(B[TG]-\d+\)`
**2. Rule ID Conversion**: `BR-S-8` → `BRS8` (remove dashes)
**3. Description Cleaning**: Remove `[BR-XX]-` prefix, normalize whitespace

## Todo List Overview

### Completed (4/28)
- ✅ Create branch
- ✅ Research sources
- ✅ Analyze structure
- ✅ Document findings

### In Progress (0/28)
- ⏸️ Ready to start implementation

### Pending (24/28)
- 📋 Setup and XML parsing (tasks 1-3)
- 📋 Data extraction (tasks 4-6)
- 📋 Testing extraction logic (task 7)
- 📋 Code generation (tasks 8-11)
- 📋 CLI implementation (tasks 12-15)
- 📋 Validation (tasks 16-18)
- 📋 Documentation and integration (tasks 19-23)

## Testing Strategy

### Phase 1: Prototype
- Generate BR-1 through BR-10
- Manual comparison with current rules.go
- Validate extraction accuracy

### Phase 2: Full Generation
- Generate all 203 rules
- Automated comparison
- Compile and run tests

### Phase 3: Integration
- Add `//go:generate` directive
- Update documentation
- CI/CD validation

## Success Metrics

- [ ] Prototype generates valid Go code
- [ ] Extraction matches official spec
- [ ] All 203 rules generated
- [ ] Existing tests pass
- [ ] Code generation is deterministic
- [ ] Documentation complete

## Development Notes

### Working Directory:
```bash
cd /home/fank/repo/einvoice
git checkout feature/autogen-rules-en16931
```

### Test Command (future):
```bash
go run ./cmd/genrules \
  -source /tmp/EN16931-CII-model-abstract.sch \
  -package rules \
  -output rules/en16931.go
```

### Validation Commands (future):
```bash
# Validate generated package compiles
go build ./rules

# Check main package can import it
go build ./...

# Run tests
go test ./...
```

## Context Preservation

This status file combined with:
1. **GENRULES_PLAN.md** - Full implementation details
2. **TODO list** - 23 tracked tasks
3. **Issue #27** - Research findings
4. **Branch commits** - Progress history

...ensures complete context can be restored at any time.

---

**Last Updated**: 2025-10-07
**Next Action**: Start task 1 - Create cmd/genrules skeleton
