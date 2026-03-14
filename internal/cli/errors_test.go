package cli

import "testing"

func TestCLIErrorErrorUsesMessage(t *testing.T) {
	t.Parallel()

	err := &cliError{Message: "boom"}
	if err.Error() != "boom" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestLooksLikeUsageError(t *testing.T) {
	t.Parallel()

	if !looksLikeUsageError(usageError("unknown_flag", "unknown flag: --bad")) {
		t.Fatal("expected usage-like error")
	}
}
