This test verifies that SilaOsaka fork blob gas calculation works correctly when
parentBaseFee is provided. It tests the SIP-7918 reserve price calculation
which requires parent.BaseFee to be properly set.

Regression test for: nil pointer dereference when parent.BaseFee was not
included in the parent header during SilaOsaka fork blob gas calculations.