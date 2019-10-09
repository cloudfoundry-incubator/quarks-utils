package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// GetNamespaceName returns a numbered namespace
func GetNamespaceName(namespaceCounter int) string {
	ns, found := os.LookupEnv("TEST_NAMESPACE")
	if !found {
		ns = "default"
	}
	return ns + "-" + strconv.Itoa(int(namespaceCounter))
}

// DumpENV executes testing/dump_env.sh to write k8s resources to files
func DumpENV(namespace string) {
	fmt.Println("Collecting debug information...")

	// try to find our dump_env script
	n := 1
	_, filename, _, _ := runtime.Caller(1)
	if idx := strings.Index(filename, "integration/"); idx >= 0 {
		n = strings.Count(filename[idx:], "/")
	}
	var dots []string
	for i := 0; i < n; i++ {
		dots = append(dots, "..")
	}
	dumpCmd := path.Join(append(dots, "testing/dump_env.sh")...)

	out, err := exec.Command(dumpCmd, namespace).CombinedOutput()
	if err != nil {
		fmt.Println("Failed to run the `dump_env.sh` script", err)
	}
	fmt.Println(string(out))
}
