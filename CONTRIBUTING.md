# Contributing to OpenCX

Thanks so much for considering to help out with OpenCX!

Please create issues for bugs and feature requests, and pull requests if you think you've solved one of the issues or have some other valuable contribution.

In any case, if there's something that you would like to see in OpenCX, **please make an issue for it!** If you want to change the code in any way, **please make a pull request!**

That being said, there are some things that would make a PR more likely to be merged:
 * You must use [gofmt](https://golang.org/cmd/gofmt/) on your code.
 * Please adhere to the [official commentary guidelines](https://golang.org/doc/effective_go.html#commentary).
 * Pull requests should be based off of, and should merge into the master branch.
 * Please run [go vet](https://golang.org/cmd/vet/) on your code before submitting a pull request. It helps the code to stay clean and consistent.

If you see anything that does not adhere to this, please make an issue for it.
Pull requests that improve anything on the [go report card](https://goreportcard.com/report/github.com/mit-dci/opencx) are welcomed!

For security-related issues, please do not file an issue, and instead contact the maintainer. A SECURITY.md file is being worked on.

## Good content for pull requests

If you don't know what to work on, here are some things that generally increase the quality of the codebase:
 * **TESTS!**
 * Anything that fixes an important issue
 * Anything that increases modularity and/or robustness in some part of the code
 * Documentation for packages and other things
 * Updates to this document that may help other contributors or developers get started
 * Anything that fixes a TODO in the code
   - If you find a TODO, feel free to create an issue for it! Sometimes they're just left there and get forgotten. Some TODOs might already be fixed, so if you create an issue and it turns out the TODO was out of date, then you probably shouldn't work on it.

## Who is the maintainer?

Currently, [rjected](https://github.com/rjected) is the maintainer.
You can contact them by [email!](mailto:dan@dancline.net)

## How can I get my pull request merged?

1 reviewer will have to review your code, and someone with write access will merge it.
Discussion on pull requests is expected to happen in the comments of the pull request.

## Some small style notes

Much of the repository has named function parameters and return variables, as well as one-liner if statements. Here's an example of all three in action:
```golang
// AvoidHelloWorld is an example function that errors out when input is "hello world"
func AvoidHelloWorld(named string) (err error) {
  if named == "hello world" {
    err = fmt.Error("The forbidden string appeared in AvoidHelloWorld")
    return
  }
  return
}

// SomethingElseUseful takes the named string, passes it through AvoidHelloWorld, and appends " world" to it.
func SomethingElseUseful(named string) (value string, err error) {
  if err = AvoidHelloWorld(named); err != nil {
    err = fmt.Errorf("Error calling AvoidHelloWorld for SomethingElseUseful: %s", err)
    return
  }

  value = named + " world"

  return
}
```

A lot of the code should be consistent in style, so you may be asked to modify code in pull requests that do not adhere to this style.

The standard libraries are pretty good, and cryptography tends to be hard, so please don't reinvent the wheel unless you have to. However, there are some things that aren't robust, compatible, or don't exist yet, and those are the exceptions to the rule.