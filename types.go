package pocketenv

import "time"

// Page holds a paginated result set.
type Page[T any] struct {
	Items []T
	Total int
}

// sandboxView is the raw JSON shape returned by the API for a sandbox.
type sandboxView struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	Description string    `json:"description"`
	Topics      []string  `json:"topics"`
	Repo        string    `json:"repo"`
	VCPUs       int       `json:"vcpus"`
	Memory      int       `json:"memory"`
	Disk        int       `json:"disk"`
	Status      string    `json:"status"`
	URI         string    `json:"uri"`
	Website     string    `json:"website"`
	Logo        string    `json:"logo"`
	Readme      string    `json:"readme"`
	CreatedAt   time.Time `json:"createdAt"`
	StartedAt   time.Time `json:"startedAt"`
}

type CreateSandboxInput struct {
	Base        string   `json:"base"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Provider    string   `json:"provider,omitempty"`
	Topics      []string `json:"topics,omitempty"`
	Repo        string   `json:"repo,omitempty"`
	VCPUs       int      `json:"vcpus,omitempty"`
	Memory      int      `json:"memory,omitempty"`
	Disk        int      `json:"disk,omitempty"`
	Readme      string   `json:"readme,omitempty"`
	KeepAlive   bool     `json:"keepAlive,omitempty"`
}

// variableView is the raw JSON shape returned by the API for an environment variable.
type variableView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	SandboxID string    `json:"sandboxId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// secretView is the raw JSON shape returned by the API for a secret.
type secretView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	SandboxID string    `json:"sandboxId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// fileView is the raw JSON shape returned by the API for a file.
type fileView struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Content   string    `json:"content"`
	SandboxID string    `json:"sandboxId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// volumeView is the raw JSON shape returned by the API for a volume.
type volumeView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	ReadOnly  bool      `json:"readOnly"`
	SandboxID string    `json:"sandboxId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// serviceView is the raw JSON shape returned by the API for a service.
type serviceView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Ports       []int  `json:"ports,omitempty"`
	Status      string `json:"status"`
	SandboxID   string `json:"sandboxId"`
	CreatedAt   string `json:"createdAt"`
}

type ExposedPort struct {
	Port        int    `json:"port"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type SshKeysView struct {
	ID         string    `json:"id"`
	PublicKey  string    `json:"publicKey"`
	PrivateKey string    `json:"privateKey"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

type AddServiceInput struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description,omitempty"`
	Ports       []int  `json:"ports,omitempty"`
}

type UpdateServiceInput struct {
	Name        string `json:"name,omitempty"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	Ports       []int  `json:"ports,omitempty"`
}
