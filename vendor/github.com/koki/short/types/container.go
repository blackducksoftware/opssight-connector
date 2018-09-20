package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/koki/short/util/floatstr"
)

type Container struct {
	Command              []string                 `json:"command,omitempty"`
	Args                 []floatstr.FloatOrString `json:"args,omitempty"`
	Env                  []Env                    `json:"env,omitempty"`
	Image                string                   `json:"image"`
	Pull                 PullPolicy               `json:"pull,omitempty"`
	OnStart              *Action                  `json:"on_start,omitempty"`
	PreStop              *Action                  `json:"pre_stop,omitempty"`
	CPU                  *CPU                     `json:"cpu,omitempty"`
	Mem                  *Mem                     `json:"mem,omitempty"`
	Name                 string                   `json:"name,omitempty"`
	AddCapabilities      []string                 `json:"cap_add,omitempty"`
	DelCapabilities      []string                 `json:"cap_drop,omitempty"`
	Privileged           *bool                    `json:"privileged,omitempty"`
	AllowEscalation      *bool                    `json:"allow_escalation,omitempty"`
	RW                   *bool                    `json:"rw,omitempty"`
	RO                   *bool                    `json:"ro,omitempty"`
	ForceNonRoot         *bool                    `json:"force_non_root,omitempty"`
	UID                  *int64                   `json:"uid,omitempty"`
	GID                  *int64                   `json:"gid,omitempty"`
	SELinux              *SELinux                 `json:"selinux,omitempty"`
	LivenessProbe        *Probe                   `json:"liveness_probe,omitempty"`
	ReadinessProbe       *Probe                   `json:"readiness_probe,omitempty"`
	Expose               []Port                   `json:"expose,omitempty"`
	Stdin                bool                     `json:"stdin,omitempty"`
	StdinOnce            bool                     `json:"stdin_once,omitempty"`
	TTY                  bool                     `json:"tty,omitempty"`
	WorkingDir           string                   `json:"wd,omitempty"`
	TerminationMsgPath   string                   `json:"termination_msg_path,omitempty"`
	TerminationMsgPolicy TerminationMessagePolicy `json:"termination_msg_policy,omitempty"`
	ContainerID          string                   `json:"container_id,omitempty"`
	ImageID              string                   `json:"image_id,omitempty"`
	Ready                bool                     `json:"ready,omitempty"`
	LastState            *ContainerState          `json:"last_state,omitempty"`
	CurrentState         *ContainerState          `json:"current_state,omitempty"`
	VolumeMounts         []VolumeMount            `json:"volume,omitempty"`
	Restarts             int32                    `json:"restarts,omitempty"`
}

type ContainerState struct {
	Waiting    *ContainerStateWaiting    `json:"waiting,omitempty"`
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
	Running    *ContainerStateRunning    `json:"running,omitempty"`
}

type ContainerStateWaiting struct {
	Reason string `json:"reason,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

type ContainerStateRunning struct {
	StartTime metav1.Time `json:"start_time,omitempty"`
}

type ContainerStateTerminated struct {
	StartTime  metav1.Time `json:"start_time,omitempty"`
	FinishTime metav1.Time `json:"finish_time,omitempty"`
	Reason     string      `json:"reason,omitempty"`
	Msg        string      `json:"msg,omitempty"`
	ExitCode   int32       `json:"exit_code,omitempty"`
	Signal     int32       `json:"signal,omitempty"`
}

type VolumeMount struct {
	MountPath   string           `json:"mount,omitempty"`
	Propagation MountPropagation `json:"propagation,omitempty"`
	Store       string           `json:"store,omitempty"`
}

type MountPropagation string

const (
	MountPropagationHostToContainer MountPropagation = "host-to-container"
	MountPropagationBidirectional   MountPropagation = "bidirectional"
	MountPropagationNone            MountPropagation = "none"
)

type TerminationMessagePolicy string

const (
	TerminationMessageReadFile              TerminationMessagePolicy = "file"
	TerminationMessageFallbackToLogsOnError TerminationMessagePolicy = "fallback-to-logs-on-error"
)

type PullPolicy string

const (
	PullAlways       PullPolicy = "always"
	PullNever        PullPolicy = "never"
	PullIfNotPresent PullPolicy = "if-not-present"
)
