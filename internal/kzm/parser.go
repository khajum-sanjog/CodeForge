package kzm

import (
	"fmt"
	"strconv"
)

// Parser parses a slice of tokens into a KZM AST.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new Parser instance.
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// Parse parses the tokens and returns the Program AST.
func (p *Parser) Parse() (*Program, error) {
	prog := &Program{
		Meta:      &Meta{},
		Triggers:  []*Trigger{},
		Variables: []*Variable{},
		Notifiers: []*Notifier{},
		Plugins:   []string{},
	}

	for p.pos < len(p.tokens) {
		tok := p.peek()

		// Stop on EOF
		if tok.Type == TokEOF {
			break
		}

		// Skip newlines at the top level
		if tok.Type == TokNewline {
			p.advance()
			continue
		}

		if tok.Type == TokKeyword {
			switch tok.Value {
			case "project":
				p.advance()
				nameTok, err := p.consume(TokString, "expected project name string")
				if err != nil {
					return nil, err
				}
				prog.Meta.Name = nameTok.Value
				p.consumeNewline()
			case "version":
				p.advance()
				verTok, err := p.consume(TokString, "expected version string")
				if err != nil {
					return nil, err
				}
				prog.Meta.Version = verTok.Value
				p.consumeNewline()
			case "description":
				p.advance()
				descTok, err := p.consume(TokString, "expected description string")
				if err != nil {
					return nil, err
				}
				prog.Meta.Description = descTok.Value
				p.consumeNewline()
			case "watch":
				// E.g. watch folder "path" (if watch folder is not matched as phrase)
				p.advance()
				srcTok, err := p.consume(TokIdentifier, "expected trigger source (github, gitlab, folder)")
				if err != nil {
					return nil, err
				}
				pathTok, err := p.consume(TokString, "expected trigger path")
				if err != nil {
					return nil, err
				}
				trigger := &Trigger{
					Source: srcTok.Value,
					Line:   tok.Line,
				}
				if srcTok.Value == "folder" {
					trigger.Path = pathTok.Value
				} else {
					trigger.Repo = pathTok.Value
				}
				// Check for optional "on branch"
				if p.peek().Type == TokKeywordPhrase && p.peek().Value == "on branch" {
					p.advance()
					branchTok, err := p.consume(TokString, "expected branch name")
					if err != nil {
						return nil, err
					}
					trigger.Branch = branchTok.Value
				}
				prog.Triggers = append(prog.Triggers, trigger)
				p.consumeNewline()
			case "every":
				p.advance()
				cronTok, err := p.consume(TokString, "expected cron schedule string")
				if err != nil {
					return nil, err
				}
				prog.Triggers = append(prog.Triggers, &Trigger{
					Source: "cron",
					Cron:   cronTok.Value,
					Line:   tok.Line,
				})
				p.consumeNewline()
			case "notify":
				p.advance()
				chanTok, err := p.consume(TokIdentifier, "expected notification channel (slack, email)")
				if err != nil {
					return nil, err
				}
				targetTok, err := p.consume(TokString, "expected notification target")
				if err != nil {
					return nil, err
				}
				prog.Notifiers = append(prog.Notifiers, &Notifier{
					Channel: chanTok.Value,
					Target:  targetTok.Value,
					Line:    tok.Line,
				})
				p.consumeNewline()
			case "plugin":
				p.advance()
				pluginTok, err := p.consume(TokString, "expected plugin name")
				if err != nil {
					return nil, err
				}
				prog.Plugins = append(prog.Plugins, pluginTok.Value)
				p.consumeNewline()
			case "environments":
				p.advance()
				_, err := p.consume(TokColon, "expected colon after environments")
				if err != nil {
					return nil, err
				}
				p.consumeNewline()
				envs, err := p.parseEnvironments()
				if err != nil {
					return nil, err
				}
				prog.Environments = envs
			case "set":
				p.advance()
				keyTok, err := p.consume(TokIdentifier, "expected variable name")
				if err != nil {
					return nil, err
				}
				_, err = p.consume(TokEquals, "expected = for variable assignment")
				if err != nil {
					return nil, err
				}
				valTok, err := p.consume(TokString, "expected variable value string")
				if err != nil {
					return nil, err
				}
				prog.Variables = append(prog.Variables, &Variable{
					Key:   keyTok.Value,
					Value: valTok.Value,
					Line:  tok.Line,
				})
				p.consumeNewline()
			default:
				return nil, fmt.Errorf("unexpected keyword %q on line %d", tok.Value, tok.Line)
			}
		} else if tok.Type == TokKeywordPhrase {
			switch tok.Value {
			case "watch github", "watch gitlab", "watch folder":
				p.advance()
				pathTok, err := p.consume(TokString, "expected repository or folder path string")
				if err != nil {
					return nil, err
				}
				source := "github"
				if tok.Value == "watch gitlab" {
					source = "gitlab"
				} else if tok.Value == "watch folder" {
					source = "folder"
				}

				trigger := &Trigger{
					Source: source,
					Line:   tok.Line,
				}
				if source == "folder" {
					trigger.Path = pathTok.Value
				} else {
					trigger.Repo = pathTok.Value
				}

				// Check for optional "on branch"
				if p.peek().Type == TokKeywordPhrase && p.peek().Value == "on branch" {
					p.advance()
					branchTok, err := p.consume(TokString, "expected branch name")
					if err != nil {
						return nil, err
					}
					trigger.Branch = branchTok.Value
				}
				prog.Triggers = append(prog.Triggers, trigger)
				p.consumeNewline()
			case "on trigger":
				p.advance()
				nameTok, err := p.consume(TokString, "expected trigger name")
				if err != nil {
					return nil, err
				}
				prog.Triggers = append(prog.Triggers, &Trigger{
					Source: "manual",
					Name:   nameTok.Value,
					Line:   tok.Line,
				})
				p.consumeNewline()
			case "use secrets from":
				p.advance()
				pathTok, err := p.consume(TokString, "expected secrets path")
				if err != nil {
					return nil, err
				}
				prog.Secrets = &Secrets{
					Path: pathTok.Value,
					Line: tok.Line,
				}
				p.consumeNewline()
			case "before deploy":
				p.advance()
				_, err := p.consume(TokColon, "expected colon after before deploy")
				if err != nil {
					return nil, err
				}
				p.consumeNewline()
				phase, err := p.parsePhase()
				if err != nil {
					return nil, err
				}
				prog.Before = phase
			case "after deploy":
				p.advance()
				_, err := p.consume(TokColon, "expected colon after after deploy")
				if err != nil {
					return nil, err
				}
				p.consumeNewline()
				phase, err := p.parsePhase()
				if err != nil {
					return nil, err
				}
				prog.After = phase
			case "deploy to":
				p.advance()
				target, err := p.parseDeployTarget()
				if err != nil {
					return nil, err
				}
				prog.Deploy = target
			case "keep last":
				p.advance()
				numTok, err := p.consume(TokNumber, "expected number for keep last")
				if err != nil {
					return nil, err
				}
				num, _ := strconv.Atoi(numTok.Value)
				prog.KeepLast = num
				p.consumeNewline()
			default:
				return nil, fmt.Errorf("unexpected keyword phrase %q on line %d", tok.Value, tok.Line)
			}
		} else {
			return nil, fmt.Errorf("unexpected token type %s (value: %q) on line %d", tok.Type, tok.Value, tok.Line)
		}
	}

	return prog, nil
}

func (p *Parser) parsePhase() (*Phase, error) {
	tok := p.peek()
	_, err := p.consume(TokIndent, "expected indented block")
	if err != nil {
		return nil, err
	}

	phase := &Phase{Line: tok.Line, Steps: []*Step{}}

	for p.pos < len(p.tokens) {
		t := p.peek()
		if t.Type == TokDedent {
			p.advance()
			break
		}
		if t.Type == TokEOF {
			return nil, fmt.Errorf("unexpected EOF inside phase block starting on line %d", tok.Line)
		}
		if t.Type == TokNewline {
			p.advance()
			continue
		}

		step, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		phase.Steps = append(phase.Steps, step)
		p.consumeNewline()
	}

	return phase, nil
}

func (p *Parser) parseStep() (*Step, error) {
	tok := p.peek()
	step := &Step{Line: tok.Line}

	if tok.Type == TokKeyword {
		switch tok.Value {
		case "run":
			p.advance()
			cmdTok, err := p.consume(TokString, "expected command string after run")
			if err != nil {
				return nil, err
			}
			step.Command = "run: " + cmdTok.Value
		case "copy":
			p.advance()
			srcTok, err := p.consume(TokString, "expected source path string")
			if err != nil {
				return nil, err
			}
			_, err = p.consume(TokIdentifier, "expected 'to'")
			if err != nil {
				return nil, err
			}
			destTok, err := p.consume(TokString, "expected destination path string")
			if err != nil {
				return nil, err
			}
			step.Command = fmt.Sprintf("copy: %s to %s", srcTok.Value, destTok.Value)
		case "set":
			p.advance()
			keyTok, err := p.consume(TokIdentifier, "expected setting name")
			if err != nil {
				return nil, err
			}
			_, err = p.consume(TokEquals, "expected = for setting")
			if err != nil {
				return nil, err
			}
			valTok, err := p.consume(TokString, "expected setting value string")
			if err != nil {
				return nil, err
			}
			step.Command = fmt.Sprintf("set: %s = %s", keyTok.Value, valTok.Value)
		case "plugin":
			p.advance()
			pluginTok, err := p.consume(TokString, "expected plugin name")
			if err != nil {
				return nil, err
			}
			step.Command = "plugin: " + pluginTok.Value
		default:
			return nil, fmt.Errorf("unexpected step action keyword %q on line %d", tok.Value, tok.Line)
		}
	} else {
		return nil, fmt.Errorf("expected step action (run, copy, set, plugin) on line %d", tok.Line)
	}

	// Parse optional modifiers
	for p.pos < len(p.tokens) {
		t := p.peek()
		if t.Type == TokNewline || t.Type == TokEOF {
			break
		}

		if t.Type == TokKeywordPhrase {
			switch t.Value {
			case "must pass":
				p.advance()
				step.MustPass = true
			case "or rollback":
				p.advance()
				step.OrRollback = true
			case "must pass or rollback":
				p.advance()
				step.MustPass = true
				step.OrRollback = true
			case "if env is":
				p.advance()
				envTok, err := p.consume(TokString, "expected environment name string")
				if err != nil {
					return nil, err
				}
				step.Condition = &Condition{
					EnvName:  "",
					EnvValue: envTok.Value,
				}
			default:
				return nil, fmt.Errorf("unexpected modifier %q on line %d", t.Value, t.Line)
			}
		} else if t.Type == TokKeyword && t.Value == "if" {
			// fallback if "if env is" isn't a single phrase for some reason
			p.advance()
			_, err := p.consume(TokIdentifier, "expected 'env'")
			if err != nil {
				return nil, err
			}
			_, err = p.consume(TokIdentifier, "expected 'is'")
			if err != nil {
				return nil, err
			}
			envTok, err := p.consume(TokString, "expected environment name string")
			if err != nil {
				return nil, err
			}
			step.Condition = &Condition{
				EnvName:  "",
				EnvValue: envTok.Value,
			}
		} else {
			break
		}
	}

	return step, nil
}

func (p *Parser) parseDeployTarget() (*DeployTarget, error) {
	tok := p.peek()
	targetTypeTok, err := p.consume(TokIdentifier, "expected deploy target type (ssh, lambda, s3, etc.)")
	if err != nil {
		return nil, err
	}
	nameTok, err := p.consume(TokString, "expected deploy target identifier/name string")
	if err != nil {
		return nil, err
	}

	target := &DeployTarget{
		Type:    targetTypeTok.Value,
		Name:    nameTok.Value,
		Options: make(map[string]string),
		EnvVars: make(map[string]string),
		Line:    tok.Line,
	}

	// Check for optional "at" modifier
	if p.peek().Type == TokIdentifier && p.peek().Value == "at" {
		p.advance()
		pathTok, err := p.consume(TokString, "expected destination path")
		if err != nil {
			return nil, err
		}
		target.Options["path"] = pathTok.Value
	}

	_, err = p.consume(TokColon, "expected colon after deploy target")
	if err != nil {
		return nil, err
	}
	p.consumeNewline()

	_, err = p.consume(TokIndent, "expected indented block for deploy configuration")
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) {
		t := p.peek()
		if t.Type == TokDedent {
			p.advance()
			break
		}
		if t.Type == TokEOF {
			return nil, fmt.Errorf("unexpected EOF inside deploy block starting on line %d", tok.Line)
		}
		if t.Type == TokNewline {
			p.advance()
			continue
		}

		if t.Type == TokKeyword && t.Value == "env" {
			p.advance()
			keyTok, err := p.consume(TokIdentifier, "expected environment variable key name")
			if err != nil {
				return nil, err
			}
			_, err = p.consume(TokEquals, "expected = for environment variable")
			if err != nil {
				return nil, err
			}
			valTok, err := p.consume(TokString, "expected environment variable value string")
			if err != nil {
				return nil, err
			}
			target.EnvVars[keyTok.Value] = valTok.Value
		} else {
			keyTok, err := p.consume(TokIdentifier, "expected configuration key")
			if err != nil {
				return nil, err
			}

			// Value can be String, Number, Boolean, or Identifier, potentially comma-separated
			var val string
			nextTok := p.peek()
			if nextTok.Type == TokString || nextTok.Type == TokNumber || nextTok.Type == TokBoolean || nextTok.Type == TokIdentifier {
				val = nextTok.Value
				p.advance()

				for p.peek().Type == TokComma {
					p.advance() // consume the comma
					nextValTok := p.peek()
					if nextValTok.Type == TokString || nextValTok.Type == TokNumber || nextValTok.Type == TokBoolean || nextValTok.Type == TokIdentifier {
						val += "," + nextValTok.Value
						p.advance()
					} else {
						return nil, fmt.Errorf("expected value after comma on line %d", nextValTok.Line)
					}
				}
			} else {
				return nil, fmt.Errorf("expected value for option %q on line %d, got %s", keyTok.Value, nextTok.Line, nextTok.Type)
			}
			target.Options[keyTok.Value] = val
		}
		p.consumeNewline()
	}

	return target, nil
}

func (p *Parser) parseEnvironments() ([]*Environment, error) {
	tok := p.peek()
	_, err := p.consume(TokIndent, "expected indented environments block")
	if err != nil {
		return nil, err
	}

	envs := []*Environment{}

	for p.pos < len(p.tokens) {
		t := p.peek()
		if t.Type == TokDedent {
			p.advance()
			break
		}
		if t.Type == TokEOF {
			return nil, fmt.Errorf("unexpected EOF inside environments block starting on line %d", tok.Line)
		}
		if t.Type == TokNewline {
			p.advance()
			continue
		}

		envNameTok, err := p.consume(TokIdentifier, "expected environment name")
		if err != nil {
			return nil, err
		}
		_, err = p.consume(TokColon, "expected colon after environment name")
		if err != nil {
			return nil, err
		}
		p.consumeNewline()

		_, err = p.consume(TokIndent, "expected indented deploy target block for environment")
		if err != nil {
			return nil, err
		}

		deployTok, err := p.consume(TokKeywordPhrase, "expected 'deploy to' inside environment")
		if err != nil {
			return nil, err
		}
		if deployTok.Value != "deploy to" {
			return nil, fmt.Errorf("expected 'deploy to' on line %d, got %q", deployTok.Line, deployTok.Value)
		}

		target, err := p.parseDeployTarget()
		if err != nil {
			return nil, err
		}

		envs = append(envs, &Environment{
			Name:   envNameTok.Value,
			Target: target,
			Line:   envNameTok.Line,
		})

		_, err = p.consume(TokDedent, "expected dedent after environment block")
		if err != nil {
			return nil, err
		}
		p.consumeNewline()
	}

	return envs, nil
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokEOF, Value: "", Line: 0, Col: 0}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) consume(expectedType TokenType, errMsg string) (Token, error) {
	tok := p.peek()
	if tok.Type != expectedType {
		return Token{}, fmt.Errorf("%s on line %d (got %s %q)", errMsg, tok.Line, tok.Type, tok.Value)
	}
	p.pos++
	return tok, nil
}

func (p *Parser) consumeNewline() {
	for p.peek().Type == TokNewline {
		p.advance()
	}
}
