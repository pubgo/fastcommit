package utils

import (
	"fmt"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/cmds/app"
	"github.com/urfave/cli/v3"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/cmds/app"
	"github.com/urfave/cli/v3"
	"golang.org/x/mod/semver"
)

func GetGitMaxTag(reg *regexp.Regexp) string {
	{
		cmdParams := []string{
			"fetch",
		}
		cmd := exec.Command("git", cmdParams...)
		fmt.Println(cmd.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		assert.Exit(err)
	}

	r, err := git.PlainOpen(".")
	if err != nil {
		assert.Exit(err)
	}
	tags, err := r.Tags()
	if err != nil {
		assert.Exit(err)
	}

	var maxVer = "v0.0.0"

	if err = tags.ForEach(func(tag *plumbing.Reference) error {
		tagName := tag.Name().Short()
		if !semver.IsValid(tagName) {
			return nil
		}
		if !reg.Match([]byte(tagName)) {
			return nil
		}
		if semver.Compare(maxVer, tagName) >= 0 {
			return nil
		}
		maxVer = tagName
		return nil
	}); err != nil {
		assert.Exit(err)
	}

	return maxVer
}

func createTag(env, tag, msg string) *cli.Command {
	return &cli.Command{
		Name: env,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "msg",
				Aliases:     []string{"m"},
				Usage:       "commit message",
				Destination: &msg,
				Required:    false,
				Value:       env,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			var err error
			regexpr := fmt.Sprintf("^v[0-9]+\\.[0-9]+\\.[0-9]+(-%s\\.[0-9]+)?$", tag)
			maxVer := GetGitMaxTag(regexp.MustCompile(regexpr))

			log.Info("max tag is ", maxVer)

			p := semver.Prerelease(maxVer)
			if p == "" {
				parts := strings.SplitN(maxVer, ".", 3)
				patch := assert.Must1(strconv.Atoi(parts[2]))
				patch += 1
				maxVer = fmt.Sprintf("%s.%s.%d-%s.1", parts[0], parts[1], patch, tag)
			} else {
				v, _ := strconv.ParseInt(strings.TrimPrefix(p, "-"+tag+"."), 10, 64)
				v += 1
				maxVer = fmt.Sprintf("%s-%s.%d", maxVer[:len(maxVer)-len(p)], tag, v)
			}

			log.Info("new tag is ", maxVer)

			{
				cmdParams := []string{
					"tag",
					"-a",
					maxVer,
					"-m",
					msg,
				}
				cmd := exec.Command("git", cmdParams...)
				fmt.Println(cmd.String())
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				assert.Exit(err)
			}

			{
				cmdParams := []string{
					"push",
					"origin",
					maxVer,
				}
				cmd := exec.Command("git", cmdParams...)
				fmt.Println(cmd.String())
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				assert.Exit(err)
			}

			return nil
		},
	}
}

func New(_ *dix.Dix) *cli.Command {
	var msg string

	return &cli.Command{
		Name: "git",
		Commands: []*cli.Command{
			createTag("demo", "demo", "demo"),
			createTag("dev", "alpha", "dev"),
			createTag("test", "beta", "test"),
			{
				Name: "release",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "msg",
						Aliases:     []string{"m"},
						Usage:       "commit message",
						Destination: &msg,
						Required:    false,
						Value:       "release",
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					var err error
					maxVer := GetGitMaxTag(regexp.MustCompile("^v[0-9]+\\.[0-9]+\\.[0-9]+$"))

					parts := strings.SplitN(maxVer, ".", 3)
					patch := assert.Must1(strconv.Atoi(parts[2]))
					patch += 1
					maxVer = fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch)

					log.Info("new tag is ", maxVer)

					{
						cmdParams := []string{
							"tag",
							"-a",
							maxVer,
							"-m",
							msg,
						}
						cmd := exec.Command("git", cmdParams...)
						fmt.Println(cmd.String())
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err = cmd.Run()
						assert.Exit(err)
					}

					{
						cmdParams := []string{
							"push",
							"origin",
							maxVer,
						}
						cmd := exec.Command("git", cmdParams...)
						fmt.Println(cmd.String())
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err = cmd.Run()
						assert.Exit(err)
					}

					return nil
				},
			},
		},
	}
}

func main() {
	di := app.NewBuilder()
	{
		di.Provide(New)
	}

	app.Run(di)
}
