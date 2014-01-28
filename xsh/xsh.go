package xsh

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/inject"
)

type Return struct {
	Stdout   string
	Stderr   string
	Exitcode int
}

func (r *Return) String() string {
	return r.Stdout
}

func (r *Return) Trim() string {
	return strings.TrimSpace(r.Stdout)
}

func Call(name string, args ...string) (ret *Return, err error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	ret = new(Return)
	//ret.Error
	err = cmd.Run()
	ret.Stdout = string(stdout.Bytes())
	ret.Stderr = string(stderr.Bytes())
	return
}

type Dir string

type Session struct {
	inj    inject.Injector
	Env    map[string]string
	Output io.Writer
}

func NewSession(a ...interface{}) *Session {
	env := map[string]string{
		"PATH": "/bin:/usr/bin:/usr/local/bin",
	}
	s := &Session{
		inj:    inject.New(),
		Output: os.Stdout,
		Env:    env,
	}
	dir := Dir("")
	args := []string{}
	s.inj.Map(env).Map(dir).Map(args)
	for _, v := range a {
		s.inj.Map(v)
	}
	return s
}

func (s *Session) Call(a ...interface{}) error {
	for _, v := range a {
		s.inj.Map(v)
	}
	values, err := s.inj.Invoke(s.invokeExec)
	if err != nil {
		return err
	}
	r := values[0]
	if r.IsNil() {
		return nil
	}
	return r.Interface().(error)
}

func (s *Session) invokeExec(cmd string, args []string, cwd Dir) error {
	envs := make([]string, 0, len(s.Env))
	for k, v := range s.Env {
		envs = append(envs, k+"="+v)
	}
	//fmt.Println(cmd, args)
	c := exec.Command(cmd, args...)
	c.Env = envs
	c.Dir = string(cwd)
	c.Stdout = s.Output
	c.Stderr = s.Output
	return c.Run()
}
