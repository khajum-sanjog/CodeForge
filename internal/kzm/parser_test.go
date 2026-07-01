package kzm

import (
	"testing"
)

func TestParserBasic(t *testing.T) {
	input := `project "My API"
version "1.0"
description "Test description"

watch github "myuser/my-api" on branch "main"
use secrets from "~/.codeforge/secrets.enc"

before deploy:
  run "npm install"
  run "npm test" must pass or rollback if env is "production"

deploy to lambda "my-api-prod":
  region "ap-south-1"
  runtime "nodejs20.x"
  memory 512
  timeout 30
  env NODE_ENV = "production"

after deploy:
  run "npm run smoke-test" must pass

keep last 3

notify slack "#deployments"
notify email "ops@company.com"
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("failed to tokenize: %v", err)
	}

	parser := NewParser(tokens)
	prog, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if prog.Meta.Name != "My API" {
		t.Errorf("expected project name 'My API', got %q", prog.Meta.Name)
	}
	if prog.Meta.Version != "1.0" {
		t.Errorf("expected project version '1.0', got %q", prog.Meta.Version)
	}
	if prog.Meta.Description != "Test description" {
		t.Errorf("expected project description, got %q", prog.Meta.Description)
	}

	if len(prog.Triggers) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(prog.Triggers))
	}
	trig := prog.Triggers[0]
	if trig.Source != "github" || trig.Repo != "myuser/my-api" || trig.Branch != "main" {
		t.Errorf("unexpected trigger: %+v", trig)
	}

	if prog.Secrets == nil || prog.Secrets.Path != "~/.codeforge/secrets.enc" {
		t.Errorf("unexpected secrets path: %+v", prog.Secrets)
	}

	if prog.Before == nil || len(prog.Before.Steps) != 2 {
		t.Fatalf("expected 2 before deploy steps, got %+v", prog.Before)
	}
	step2 := prog.Before.Steps[1]
	if step2.Command != "run: npm test" || !step2.MustPass || !step2.OrRollback || step2.Condition == nil || step2.Condition.EnvValue != "production" {
		t.Errorf("unexpected step2 state: %+v", step2)
	}

	if prog.Deploy == nil || prog.Deploy.Type != "lambda" || prog.Deploy.Name != "my-api-prod" {
		t.Fatalf("unexpected deploy target: %+v", prog.Deploy)
	}
	if prog.Deploy.Options["region"] != "ap-south-1" || prog.Deploy.Options["memory"] != "512" {
		t.Errorf("unexpected options: %+v", prog.Deploy.Options)
	}
	if prog.Deploy.EnvVars["NODE_ENV"] != "production" {
		t.Errorf("unexpected env: %+v", prog.Deploy.EnvVars)
	}

	if prog.After == nil || len(prog.After.Steps) != 1 {
		t.Fatalf("expected 1 after deploy step, got %+v", prog.After)
	}

	if prog.KeepLast != 3 {
		t.Errorf("expected keep last 3, got %d", prog.KeepLast)
	}

	if len(prog.Notifiers) != 2 {
		t.Errorf("expected 2 notifiers, got %d", len(prog.Notifiers))
	}
}

func TestParserEnvironments(t *testing.T) {
	input := `project "My App"
environments:
  production:
    deploy to lambda "my-lambda-prod":
      region "us-east-1"
  staging:
    deploy to local "local-deploy":
      path "/tmp/deploy"
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("failed to tokenize: %v", err)
	}

	parser := NewParser(tokens)
	prog, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(prog.Environments) != 2 {
		t.Fatalf("expected 2 environments, got %d", len(prog.Environments))
	}

	env1 := prog.Environments[0]
	if env1.Name != "production" || env1.Target.Type != "lambda" || env1.Target.Options["region"] != "us-east-1" {
		t.Errorf("unexpected env1: %+v", env1)
	}

	env2 := prog.Environments[1]
	if env2.Name != "staging" || env2.Target.Type != "local" || env2.Target.Options["path"] != "/tmp/deploy" {
		t.Errorf("unexpected env2: %+v", env2)
	}
}
