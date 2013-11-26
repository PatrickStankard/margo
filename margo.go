package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
)

type Config struct {
	Jobs []Jobs
}

type Jobs struct {
	Name  string
	Tasks []Task
}

type Task struct {
	Hosts    []string
	Key      string
	Commands []Command
	Options  TaskOptions
}

type Command struct {
	User, Exec string
}

type TaskOptions struct {
	Hosts    HostsOptions
	Commands CommandsOptions
}

type HostsOptions struct {
	Sync bool
}

type CommandsOptions struct {
	Sync bool
}

type Keychain struct {
	Keys []ssh.Signer
}

func errorHandler(x *error) {
	if *x != nil {
		panic(*x)
	}
}

func (k *Keychain) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.Keys) {
		return nil, nil
	}

	return k.Keys[i].PublicKey(), nil
}

func (k *Keychain) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	return k.Keys[i].Sign(rand, data)
}

func (k *Keychain) LoadPEM(file string) (err error) {
	buf, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	key, err := ssh.ParsePrivateKey(buf)

	if err != nil {
		return err
	}

	k.Keys = append(k.Keys, key)

	return nil
}

func createClient(key *string, address *string, user *string) (client *ssh.ClientConn, err error) {
	k := new(Keychain)

	err = k.LoadPEM(*key)

	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthKeyring(k),
		},
	}

	client, err = ssh.Dial("tcp", *address, config)

	if err != nil {
		return nil, err
	}

	return client, nil
}

func parseConfig(path *string) (config *Config, err error) {
	file, err := ioutil.ReadFile(*path)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &config)

	return config, nil
}

func runTask(task *Task, tg *sync.WaitGroup) {
	var hg sync.WaitGroup

	hosts := task.Hosts
	key := task.Key
	commands := task.Commands
	options := task.Options

	for m, n := 0, len(hosts); m < n; m++ {
		hg.Add(1)

		if options.Hosts.Sync == true {
			executeHostCommands(&hosts[m], &key, &options, commands[:], &hg)
		} else {
			go executeHostCommands(&hosts[m], &key, &options, commands[:], &hg)
		}
	}

	hg.Wait()
	tg.Done()
}

func executeHostCommands(host, key *string, options *TaskOptions, commands []Command, hg *sync.WaitGroup) {
	var cg sync.WaitGroup

	for o, p := 0, len(commands); o < p; o++ {
		cg.Add(1)

		user := commands[o].User
		exec := commands[o].Exec

		if options.Commands.Sync == true {
			executeCommand(&user, &exec, host, key, &cg)
		} else {
			go executeCommand(&user, &exec, host, key, &cg)
		}
	}

	cg.Wait()
	hg.Done()
}

func executeCommand(user, exec, host, key *string, cg *sync.WaitGroup) {
	client, err := createClient(key, host, user)
	errorHandler(&err)

	defer client.Close()

	session, err := client.NewSession()
	errorHandler(&err)

	defer session.Close()

	var buf bytes.Buffer

	session.Stdout = &buf

	err = session.Run(*exec)
	errorHandler(&err)

	fmt.Println(*exec + " - " + *user + "@" + *host + ":")
	fmt.Println(buf.String())

	cg.Done()
}

func main() {
	var configFile, jobToRun string

	flag.StringVar(&configFile, "config", "", "Location of config.json")
	flag.StringVar(&jobToRun, "job", "", "Name of the job to run")

	flag.Parse()

	if configFile == "" {
		err := errors.New("Missing config flag (ex: -config \"/tmp/config.json\")")
		errorHandler(&err)
	}

	if jobToRun == "" {
		err := errors.New("Missing job flag (ex: -job \"My Job\")")
		errorHandler(&err)
	}

	config, err := parseConfig(&configFile)
	errorHandler(&err)

	var tg sync.WaitGroup

	for i, j := 0, len(config.Jobs); i < j; i++ {
		job := config.Jobs[i]
		name := job.Name

		if name == jobToRun {
			tasks := job.Tasks

			for k, l := 0, len(tasks); k < l; k++ {
				tg.Add(1)

				go runTask(&tasks[k], &tg)
			}
		}
	}

	tg.Wait()
}
