package names

// SpecIndex return the job sepc index.
// We use a very large value as a maximum number of replicas per instance group, per AZ
// We do this in lieu of using the actual replica count, which would cause pods to always restart
func SpecIndex(azIndex int, podOrdinal int) int {
	if azIndex < 1 {
		azIndex = 1
	}

	return (azIndex-1)*10000 + podOrdinal
}
