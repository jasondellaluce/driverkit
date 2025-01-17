//go:build !race
// +build !race

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"gotest.tools/assert"
)

type expect struct {
	err            string
	out            string
	fmtRuntimeArch bool
}

type testCase struct {
	descr  string
	env    map[string]string
	args   []string
	expect expect
}

var tests = []testCase{
	{
		args: []string{"help"},
		expect: expect{
			out:            "testdata/help.txt",
			fmtRuntimeArch: true,
		},
	},
	{
		args: []string{"-h"},
		expect: expect{
			out:            "testdata/help-flag.txt",
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "empty",
		args:  []string{},
		expect: expect{
			out:            "testdata/autohelp.txt",
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "invalid/processor",
		args:  []string{"abc"},
		expect: expect{
			out:            "testdata/non-existent-processor.txt",
			err:            `invalid argument "abc" for "driverkit"`,
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "invalid/config/proxy",
		args: []string{
			"--proxy",
			"wrong",
		},
		expect: expect{
			out:            "testdata/invalid-proxyconfig.txt",
			err:            "exiting for validation errors",
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "docker/all-flags",
		args: []string{
			"docker",
			"--kernelrelease",
			"4.15.0-1057-aws",
			"--kernelversion",
			"59",
			"--kernelurls",
			"https://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-aws/linux-aws-headers-4.15.0-1057_4.15.0-1057.59_all.deb",
			"--kernelurls",
			"https://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-aws/linux-headers-4.15.0-1057-aws_4.15.0-1057.59_amd64.deb",
			"--target",
			"ubuntu-aws",
			"--output-module",
			"/tmp/falco-ubuntu-aws.ko",
		},
		expect: expect{
			out: "testdata/docker-with-flags.txt",
		},
	},
	{
		descr: "docker/empty",
		args:  []string{"docker"},
		expect: expect{
			err:            "exiting for validation errors",
			out:            "testdata/dockernoopts.txt",
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "docker/all-flags-debug",
		args: []string{
			"docker",
			"--kernelrelease",
			"4.15.0-1057-aws",
			"--kernelversion",
			"59",
			"--kernelurls",
			"https://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-aws/linux-aws-headers-4.15.0-1057_4.15.0-1057.59_all.deb",
			"--kernelurls",
			"https://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-aws/linux-headers-4.15.0-1057-aws_4.15.0-1057.59_amd64.deb",
			"--target",
			"ubuntu-aws",
			"--output-module",
			"/tmp/falco-ubuntu-aws.ko",
			"--loglevel",
			"debug",
		},
		expect: expect{
			out:            "testdata/docker-with-flags-debug.txt",
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "docker/merge-from-env",
		env: map[string]string{
			"DRIVERKIT_KERNELVERSION": "59",
			"DRIVERKIT_OUTPUT_MODULE": "/tmp/falco-ubuntu-aws.ko",
		},
		args: []string{
			"docker",
			"--kernelrelease",
			"4.15.0-1057-aws",
			"--kernelurls",
			"https://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-aws/linux-aws-headers-4.15.0-1057_4.15.0-1057.59_all.deb",
			"--kernelurls",
			"https://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-aws/linux-headers-4.15.0-1057-aws_4.15.0-1057.59_amd64.deb",
			"-t",
			"ubuntu-aws",
			"--loglevel",
			"debug",
		},
		expect: expect{
			out:            "testdata/docker-with-flags-debug.txt",
			fmtRuntimeArch: true,
		},
	},
	{
		descr: "docker/from-config-file",
		args: []string{
			"docker",
			"-c",
			"testdata/configs/1.yaml",
			"--loglevel",
			"debug",
		},
		expect: expect{
			out:            "testdata/docker-from-config-debug.txt",
			fmtRuntimeArch: true,
		},
	},
	{
	    descr: "docker/from-config-file-using-urls",
        args: []string{
            "docker",
            "-c",
            "testdata/configs/2.yaml",
            "--loglevel",
            "debug",
        },
        expect: expect{
            out:            "testdata/docker-override-urls-from-config-debug.txt",
            fmtRuntimeArch: true,
        },
    },
	{
		descr: "docker/override-from-config-file",
		env: map[string]string{
			"DRIVERKIT_KERNELVERSION": "229",
			"DRIVERKIT_OUTPUT_MODULE": "/tmp/override.ko",
		},
		args: []string{
			"docker",
			"-c",
			"testdata/configs/1.yaml",
			"--loglevel",
			"debug",
		},
		expect: expect{
			out:            "testdata/docker-override-from-config-debug.txt",
			fmtRuntimeArch: true,
		},
	},
	{
	    descr: "docker/build-related-target-azure",
	    args: []string{
	        "docker",
	        "--kernelrelease",
	        "4.15.0-1057-azure",
	        "--kernelversion",
	        "62",
	        "--kernelurls",
	        "http://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-azure/linux-azure-headers-4.15.0-1057_4.15.0-1057.62_all.deb",
	        "--kernelurls",
	        "http://mirrors.edge.kernel.org/ubuntu/pool/main/l/linux-azure/linux-headers-4.15.0-1057-azure_4.15.0-1057.62_amd64.deb",
	        "--target",
	        "ubuntu-aws",
	        "--output-module",
	        "/tmp/falco-ubuntu-azure.ko",
	        "--loglevel",
	        "debug",
        },
        expect: expect{
            out:            "testdata/docker-related-target-debug.txt",
            fmtRuntimeArch: true,
        },
    },
    {
        descr: "docker/build-target-check-validation-redhat",
        args: []string{
            "docker",
            "--kernelrelease",
            "4.18.0-348.el8.x86_64",
            "--target",
            "redhat",
            "--output-module",
            "/tmp/falco-redhat.ko",
            "--loglevel",
            "debug",
        },
        expect: expect{
            out:            "testdata/docker-target-redhat-validation-error-debug.txt",
            err:            "exiting for validation errors",
            fmtRuntimeArch: true,
        },
    },
	{
		descr: "complete/docker/targets",
		args: []string{
			"__complete",
			"docker",
			"--target",
			"ENTER",
		},
		expect: expect{
			out: "testdata/completion-targets.txt",
		},
	},
	{
		descr: "complete/kubernetes-alias/targets",
		args: []string{
			"__complete",
			"k8s",
			"--target",
			"ENTER",
		},
		expect: expect{
			out: "testdata/completion-targets.txt",
		},
	},
	{
		descr: "complete/kubernetes/targets",
		args: []string{
			"__complete",
			"kubernetes",
			"--target",
			"ENTER",
		},
		expect: expect{
			out: "testdata/completion-targets.txt",
		},
	},
	{
		descr: "completion/empty",
		args: []string{
			"completion",
		},
		expect: expect{
			out: "testdata/completion-noargs.txt",
		},
	},
	{
		descr: "completion/help",
		args: []string{
			"completion",
			"help",
		},
		expect: expect{
			out: "testdata/completion-noargs.txt",
		},
	},
	{
		descr: "completion/help-short-flag",
		args: []string{
			"completion",
			"-h",
		},
		expect: expect{
			out: "testdata/completion-noargs.txt",
		},
	},
}

func run(t *testing.T, test testCase) {
	// Setup
	c := NewRootCmd()
	b := bytes.NewBufferString("")
	c.SetOutput(b)
	if len(test.args) == 0 || (test.args[0] != "__complete" && test.args[0] != "__completeNoDesc" && test.args[0] != "help" && test.args[0] != "completion") {
		test.args = append(test.args, "--dryrun")
	}
	c.SetArgs(test.args)
	for k, v := range test.env {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("error setting env variables: %v", err)
		}
	}
	// Test
	err := c.Execute()
	if err != nil {
		if test.expect.err == "" {
			t.Fatalf("error executing CLI: %v", err)
		} else {
			assert.Error(t, err, test.expect.err)
		}
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatalf("error reading CLI output: %v", err)
	}
	res := stripansi.Strip(string(out))
	assert.Equal(t, test.expect.out, res)
	// Teardown
	for k := range test.env {
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("error tearing down: %v", err)
		}
	}
}

func TestCLI(t *testing.T) {
	for _, test := range tests {
		descr := test.descr
		if descr == "" {
			if test.expect.out == "" {
				t.Fatal("malformed test case: missing both descr and expect.out fields")
			}
			test.descr = strings.TrimSuffix(filepath.Base(test.expect.out), ".txt")
		}
		if test.expect.out != "" {
			out, err := ioutil.ReadFile(test.expect.out)
			if err != nil {
				t.Fatalf("output fixture not found: %v", err)
			}
			if !test.expect.fmtRuntimeArch {
				test.expect.out = string(out)
			} else {
				test.expect.out = fmt.Sprintf(string(out), runtime.GOARCH)
			}
		}

		t.Run(test.descr, func(t *testing.T) {
			run(t, test)
		})
	}
}
