package v1

import (
	fmt "fmt"
	"strconv"
	"strings"
)

// TODO(mitchdraft) - this should error check for length
// TODO(mitchdraft) - this should include process id so that we can have multiple debuggers on the same container
// TODO(mitchdraft) - generated identifier should be applied as a label - the point is to make the resource queriable
func GenDebugAttachmentName(pod, container string) string {
	return fmt.Sprintf("%v-%v", pod, container)
}

func (m *DebugAttachment) GetPortFromDebugServerAddress() (int, error) {
	if m.DebugServerAddress == "" {
		return 0, fmt.Errorf("No debug server address specified on debug attachment %v in namespace %v", m.Metadata.Name, m.Metadata.Namespace)
	}
	parts := strings.Split(m.DebugServerAddress, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("Invalid debug server address (%v) specified on debug attachment %v in namespace %v", m.DebugServerAddress, m.Metadata.Name, m.Metadata.Namespace)
	}
	return strconv.Atoi(parts[1])
}
