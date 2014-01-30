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
	Stdout string
	Stderr string
}

func (r *Return) String() string {
	return r.Stdout
}

func (r *Return) Trim() string {
	return strings.TrimSpace(r.Stdout)
}

func Capture(a ...interface{}) (ret *Return, err error) {
	s := NewSession(a...)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	s.Stdout, s.Stderr = stdout, stderr
	err = s.Call()

	ret = new(Return)
	ret.Stdout = string(stdout.Bytes())
	ret.Stderr = string(stderr.Bytes())
	return
}

type Dir string

type Session struct {
	inj    inject.Injector
	Env    map[string]string
	Stdout io.Writer
	Stderr io.Writer
}

func NewSession(a ...interface{}) *Session {
	env := map[string]string{
		"PATH": "/bin:/usr/bin:/usr/local/bin",
	}
	s := &Session{
		inj:    inject.New(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
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
	c.Stdout = s.Stdout
	c.Stderr = s.Stderr
	return c.Run()
}
