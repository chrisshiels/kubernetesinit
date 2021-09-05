// 'kubernetesinit.go'.
// Chris Shiels.


package main


import (
  "errors"
  "flag"
  "fmt"
  "io"
  "io/ioutil"
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "time"

  "gopkg.in/yaml.v2"
)


type KubernetesInit struct {
  APIVersion string   `yaml:"apiVersion"`
  Kind       string   `yaml:"kind"`
  Namespace  string   `yaml:"namespace"`
  Type       string   `yaml:"type"`
  RetryLimit int      `yaml:"retryLimit"`
  Wait       []string `yaml:"wait"`
}


func readkubernetesinit(path string) (*KubernetesInit, error) {
  bytes, err := ioutil.ReadFile(path)
  if err != nil {
    return nil, err
  }

  var kubernetesinit KubernetesInit
  kubernetesinit.Type = "kustomize"
  kubernetesinit.RetryLimit = 3

  err = yaml.Unmarshal(bytes, &kubernetesinit)
  if err != nil {
    return nil, err
  }

  return &kubernetesinit, nil
}


func command(dryrun bool,
             command string,
             stdin io.Reader,
             stdout io.Writer,
             stderr io.Writer) error {
  fmt.Printf("Run:  %s\n", command)
  if dryrun {
    return nil
  }
  split := strings.Split(command, " ")
  cmd := exec.Command(split[0], split[1:]...)
  cmd.Stdin = stdin
  cmd.Stdout = stdout
  cmd.Stderr = stderr
  return cmd.Run()
}


func repeatcommand(n int, f func () error) error {
  const delay = 5
  var err error
  for i := 1; i <= n; i++ {
    err = f()
    if err == nil {
      break
    }

    var exiterror *exec.ExitError
    exiterror, ok := err.(*exec.ExitError)
    if ok {
      fmt.Printf("Exit status %d.  Retrying in %ds.\n",
                 exiterror.ExitCode(), delay)
    } else {
      fmt.Printf("Retrying in %ds.\n", delay)
    }

    time.Sleep(time.Duration(delay) * time.Second)
  }
  return err
}


func buildkubectlcommand(kubectloptions string) string {
  if kubectloptions != "" {
    return fmt.Sprintf("kubectl %s", kubectloptions)
  }
  return "kubectl"
}


func kubectlapplycommand(dryrun bool,
                         kubectlcommandstring string,
                         filename string) func () error {
  return func() error {
    file, err := os.Open(filename)
    if err != nil {
      return err
    }

    cmderr := command(dryrun, kubectlcommandstring, file, os.Stdout, os.Stderr)

    err = file.Close()
    if err != nil {
      return err
    }

    return cmderr
  }
}


func kubectlwaitcommand(dryrun bool,
                        kubectlcommandstring string) func () error {
  return func() error {
    return command(dryrun, kubectlcommandstring, nil, os.Stdout, os.Stderr)
  }
}


func filterstrings(ss []string,
                   f func(s string) (ok bool, err error)) (ss1 []string,
                                                           err error) {
  for _, s := range ss {
    ok, err := f(s)
    if err != nil {
      return nil, err
    }
    if ok {
      ss1 = append(ss1, s)
    }
  }

  return ss1, err
}


func validenvironmentfilter(environment string) func(s string) (ok bool,
                                                                err error) {
  return func(directory string) (ok bool, err error) {
    environmentdirectory := filepath.Join(directory, environment)
    _, err = os.Stat(environmentdirectory)
    if err != nil {
      return false, nil
    }

    kubernetesinityaml := filepath.Join(directory, "kubernetesinit.yaml")
    _, err = os.Stat(kubernetesinityaml)
    if err != nil {
      return false, nil
    }

    return true, nil
  }
}


func processsubdirectory(dryrun bool,
                         kubectloptions string,
                         directory string,
                         environment string) error {
  kubernetesinityaml := filepath.Join(directory, "kubernetesinit.yaml")
  kubernetesinit, err := readkubernetesinit(kubernetesinityaml)
  if err != nil {
    return err
  }

  // kustomize build.
  tempfile, err := ioutil.TempFile("", "kustomizeinit-kustomize")
  if err != nil {
    return err
  }

  defer os.Remove(tempfile.Name())

  kustomizecommandstring := fmt.Sprintf("kustomize build %s",
                                        filepath.Join(directory, environment))
  cmderr := command(dryrun, kustomizecommandstring, nil, tempfile, os.Stderr)

  err = tempfile.Close()
  if err != nil {
    return err
  }

  if cmderr != nil {
    return cmderr
  }

  // kubectl apply.
  kubectlcommandstring := fmt.Sprintf("%s apply -f -",
                                      buildkubectlcommand(kubectloptions))
  err = repeatcommand(kubernetesinit.RetryLimit,
                      kubectlapplycommand(dryrun,
                                          kubectlcommandstring,
                                          tempfile.Name()))
  if err != nil {
    return err
  }

  if !dryrun {
    time.Sleep(1 * time.Second)
  }

  // kubectl wait.
  for _, wait := range kubernetesinit.Wait {
    kubectlcommandstring := fmt.Sprintf("%s -n %s %s",
                                        buildkubectlcommand(kubectloptions),
                                        kubernetesinit.Namespace,
                                        wait)
    err = repeatcommand(kubernetesinit.RetryLimit,
                        kubectlwaitcommand(dryrun,
                                           kubectlcommandstring))
    if err != nil {
      return err
    }
  }

  return nil
}


func processdirectory(dryrun bool,
                      kubectloptions string,
                      directory string,
                      environment string) error {
  _, err := exec.LookPath("kustomize")
  if err != nil {
    return err
  }

  _, err = exec.LookPath("kubectl")
  if err != nil {
    return err
  }

  subdirectories, err := filepath.Glob(filepath.Join(directory, "*"))
  if err != nil {
    return err
  }

  subdirectories, err = filterstrings(subdirectories,
                                      validenvironmentfilter(environment))
  if err != nil {
    return err
  }

  for i, subdirectory := range subdirectories {
    if i > 0 {
      fmt.Printf("\n\n")
    }
    err = processsubdirectory(dryrun,
                              kubectloptions,
                              subdirectory,
                              environment)
    if err != nil {
      return err
    }
  }
  return nil
}


func parseargs(args []string) (func() error, error) {
  var flagset = flag.NewFlagSet("kubernetesinit", flag.ContinueOnError)

  dryrun := flagset.Bool("dryrun", false,
                         "Dry-run (default false)")
  kubectloptions := flagset.String("kubectloptions", "",
                                   "Kubectl options (default \"\")")

  directory := flagset.String("directory", "",
                              "Directory (required)")
  environment := flagset.String("environment", "",
                                "Environment (required)")

  err := flagset.Parse(args[1:])

  switch {
  case err == flag.ErrHelp:
    return func() error {
      return nil
    }, nil
  case err != nil:
    return nil, err
  case *directory == "" || *environment == "" || len(flagset.Args()) != 0:
    flagset.Usage()
    return nil, errors.New("Missing arguments")
  default:
    return func() error {
      return processdirectory(*dryrun,
                              *kubectloptions,
                              *directory,
                              *environment)
    }, nil
  }
}


func main() {
  f, err := parseargs(os.Args)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error:  %v\n", err)
    os.Exit(1)
  }

  err = f()
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error:  %v\n", err)
    os.Exit(1)
  }

  os.Exit(0)
}
