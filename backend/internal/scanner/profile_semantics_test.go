package scanner

import "testing"

func TestParseProfileArgsRejectsInconsistentScans(t *testing.T) {
	for _, args := range []string{
		"-sS -sA", "-sT -sS", "-p80 -p443", "-F -p80", "--top-ports 0", "--top-ports 65536",
		"-p65536", "-p100-1", "-p80,", "-sn -p80", "-sn -sV", "-sn -sS", "-sA --open", "-sA -sV", "-sA -O",
	} {
		t.Run(args, func(t *testing.T) {
			if _, err := ParseProfileArgs(args); err == nil {
				t.Fatalf("expected invalid configuration to be rejected: %s", args)
			}
		})
	}
}

func TestParseProfileArgsAcceptsCompatibleScans(t *testing.T) {
	for _, args := range []string{"-sS -sU -p T:80,443,U:53,161 -T3", "-sA -Pn -T3", "-sn -PE -PS22,443", "-p1-65535 -sV", "--top-ports 100 -sV"} {
		if _, err := ParseProfileArgs(args); err != nil {
			t.Errorf("valid configuration %q rejected: %v", args, err)
		}
	}
}
