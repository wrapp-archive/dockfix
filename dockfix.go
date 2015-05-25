package dockfix

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/fsouza/go-dockerclient"
	"github.com/wrapp/env"
)

var dockerURL string

func OpenClient() (*docker.Client, error) {

	dockerURL = env.Default("DOCKER_HOST", "unix:///var/run/docker.sock")
	dockerCertPath := env.Default("DOCKER_CERT_PATH", "")

	if dockerCertPath != "" {
		ca := path.Join(dockerCertPath, "ca.pem")
		key := path.Join(dockerCertPath, "key.pem")
		cert := path.Join(dockerCertPath, "cert.pem")

		return docker.NewTLSClient(dockerURL, cert, key, ca)
	}
	return docker.NewClient(dockerURL)
}

func StartContainer(name, baseImage string) *docker.Container {
	containerFileName := baseImage + ".container"

	cid, _ := ioutil.ReadFile(containerFileName)
	fmt.Println("Container: ", string(cid))

	var containerID string

	log.Println("dockerURL", dockerURL)
	dc, err := OpenClient()

	if err != nil {
		panic(err)
	}
	if len(cid) != 0 {
		fmt.Println("Container exists")
		containerID = string(cid)
	} else {
		log.Print("Creating new container for ", baseImage)
		cont, err := dc.CreateContainer(
			docker.CreateContainerOptions{
				Config: &docker.Config{
					Image: baseImage,
				},
			},
		)
		if err != nil {
			fmt.Println(err)
		}
		containerID = cont.ID
	}

	ioutil.WriteFile(containerFileName, []byte(containerID), 0644)

	// Start container
	hc := docker.HostConfig{
		PublishAllPorts: true,
	}
	dc.StartContainer(containerID, &hc)
	cont, err := dc.InspectContainer(containerID)
	if err != nil {
		log.Println(err)
	}
	return cont
}

func StopContainer(c *docker.Container) {
	dc, _ := OpenClient()
	dc.KillContainer(docker.KillContainerOptions{
		ID: c.ID,
	})
}
