package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new KZM pipeline configuration interactively",
	Long:  `Runs a step-by-step console wizard to configure a project trigger, deployment adapter target, testing script, and notification alerts.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		color.Cyan("Welcome to CodeForge Project Wizard!\n")
		color.HiBlack("Let's configure your CI/CD pipeline step-by-step.\n")

		// Step 1: Project Metadata
		projectName := prompt(reader, "Project name", "My Awesome App")
		projectVer := prompt(reader, "Version", "1.0")
		projectDesc := prompt(reader, "Description (optional)", "A Go application deployed automatically")

		// Step 2: Trigger Source
		color.Cyan("\nStep 2: Configuration Trigger Source")
		fmt.Println("1) GitHub Push")
		fmt.Println("2) GitLab Push")
		fmt.Println("3) Local Folder Watcher")
		fmt.Println("4) Cron Scheduler")
		fmt.Println("5) Manual Trigger only")
		trigChoice := prompt(reader, "Select trigger type [1-5]", "1")

		var triggerLines []string
		switch trigChoice {
		case "1":
			repo := prompt(reader, "GitHub repository (e.g. user/repo)", "myuser/my-app")
			branch := prompt(reader, "Target branch name", "main")
			triggerLines = append(triggerLines, fmt.Sprintf("watch github %q on branch %q", repo, branch))
		case "2":
			repo := prompt(reader, "GitLab repository (e.g. user/repo)", "myuser/my-app")
			branch := prompt(reader, "Target branch name", "main")
			triggerLines = append(triggerLines, fmt.Sprintf("watch gitlab %q on branch %q", repo, branch))
		case "3":
			path := prompt(reader, "Local path to watch", "./src")
			triggerLines = append(triggerLines, fmt.Sprintf("watch folder %q", path))
		case "4":
			cronVal := prompt(reader, "Cron schedule (e.g. 'hourly', '5m', '0 0 * * *')", "daily")
			triggerLines = append(triggerLines, fmt.Sprintf("every %q", cronVal))
		default:
			triggerLines = append(triggerLines, "on trigger \"deploy\"")
		}

		// Step 3: Deployment Target
		color.Cyan("\nStep 3: Deployment Target Adapter")
		fmt.Println("1) Local Copy")
		fmt.Println("2) SSH Server (SFTP + Restart Command)")
		fmt.Println("3) AWS Lambda Function")
		fmt.Println("4) cPanel Server")
		fmt.Println("5) AWS S3 Bucket")
		fmt.Println("6) Docker Remote Host")
		fmt.Println("7) VPS Virtual Machine")
		deployChoice := prompt(reader, "Select deploy target [1-7]", "1")

		var targetLines []string
		switch deployChoice {
		case "1":
			path := prompt(reader, "Local destination directory path", "/var/www/html")
			targetLines = append(targetLines, "deploy to local \"local-deployment\":", fmt.Sprintf("  path %q", path))
		case "2":
			server := prompt(reader, "SSH target address (e.g. user@host)", "ubuntu@192.168.1.10")
			path := prompt(reader, "Remote directory path", "/var/www/my-app")
			key := prompt(reader, "Private SSH key file path", "~/.ssh/id_rsa")
			restart := prompt(reader, "Process restart command (e.g. pm2 restart app)", "pm2 restart my-app")
			targetLines = append(targetLines, fmt.Sprintf("deploy to ssh %q at %q:", server, path), fmt.Sprintf("  key     %q", key), fmt.Sprintf("  restart %q", restart))
		case "3":
			funcName := prompt(reader, "Lambda function name", "my-serverless-api")
			region := prompt(reader, "AWS region", "us-east-1")
			runtimeVal := prompt(reader, "Lambda execution runtime", "nodejs20.x")
			targetLines = append(targetLines, fmt.Sprintf("deploy to lambda %q:", funcName), fmt.Sprintf("  region  %q", region), fmt.Sprintf("  runtime %q", runtimeVal), "  memory  512", "  timeout 30")
		case "4":
			server := prompt(reader, "cPanel FTP/SFTP server domain", "ftp.myhost.com")
			user := prompt(reader, "cPanel username", "mycpaneluser")
			path := prompt(reader, "Remote directory path", "/public_html")
			targetLines = append(targetLines, fmt.Sprintf("deploy to cpanel %q at %q:", server, path), fmt.Sprintf("  user    %q", user), "  exclude \".env,vendor,.git\"")
		case "5":
			bucket := prompt(reader, "S3 bucket name", "my-static-website-bucket")
			folder := prompt(reader, "Build files directory", "./dist")
			region := prompt(reader, "AWS Region", "us-east-1")
			targetLines = append(targetLines, fmt.Sprintf("deploy to s3 %q:", bucket), fmt.Sprintf("  folder     %q", folder), fmt.Sprintf("  region     %q", region), "  public     yes", "  invalidate cloudfront")
		case "6":
			image := prompt(reader, "Docker image tag name", "myuser/my-app:latest")
			server := prompt(reader, "SSH remote server", "ubuntu@myserver.com")
			port := prompt(reader, "Exposed port map", "8080")
			targetLines = append(targetLines, fmt.Sprintf("deploy to docker %q:", "my-container-app"), fmt.Sprintf("  image   %q", image), fmt.Sprintf("  server  %q", server), fmt.Sprintf("  port    %s", port), "  restart always")
		case "7":
			server := prompt(reader, "VPS host address (e.g. root@vps.com)", "root@myvps.com")
			path := prompt(reader, "Target directory path", "/var/www/html")
			restart := prompt(reader, "Restart command (e.g. systemctl restart nginx)", "systemctl restart nginx")
			targetLines = append(targetLines, fmt.Sprintf("deploy to vps %q:", server), fmt.Sprintf("  path    %q", path), fmt.Sprintf("  restart %q", restart), "  git     yes")
		}

		// Step 4: Unit testing before deploy
		color.Cyan("\nStep 4: Testing & Validation")
		testCmd := prompt(reader, "Command to run before deploy (leave empty to skip)", "npm test")

		// Step 5: Notifications
		color.Cyan("\nStep 5: Deployment Notification Alerts")
		slackWebhook := prompt(reader, "Slack webhook URL (leave empty to skip)", "")
		emailAddr := prompt(reader, "Notification recipient email address (leave empty to skip)", "")

		// Compile .kzm file
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("project %q\n", projectName))
		sb.WriteString(fmt.Sprintf("version %q\n", projectVer))
		if projectDesc != "" {
			sb.WriteString(fmt.Sprintf("description %q\n", projectDesc))
		}
		sb.WriteString("\n")

		for _, tl := range triggerLines {
			sb.WriteString(tl + "\n")
		}
		sb.WriteString("\n")

		if testCmd != "" {
			sb.WriteString("before deploy:\n")
			sb.WriteString(fmt.Sprintf("  run %q must pass or rollback\n\n", testCmd))
		}

		for _, tl := range targetLines {
			sb.WriteString(tl + "\n")
		}
		sb.WriteString("\n")

		if slackWebhook != "" {
			sb.WriteString(fmt.Sprintf("notify slack %q\n", slackWebhook))
		}
		if emailAddr != "" {
			sb.WriteString(fmt.Sprintf("notify email %q\n", emailAddr))
		}

		// Output file
		kzmContent := sb.String()
		fileName := sanitizeFilename(projectName) + ".kzm"

		color.Cyan("\nPreview of generated %s config:\n", fileName)
		fmt.Println(kzmContent)

		confirm := prompt(reader, "Save configuration to current folder? [yes/no]", "yes")
		if strings.ToLower(confirm) == "yes" || strings.ToLower(confirm) == "y" {
			err := os.WriteFile(fileName, []byte(kzmContent), 0644)
			if err != nil {
				color.Red("Error: failed to save config file: %v\n", err)
				return
			}
			color.Green("CodeForge ✓  Config saved to %s", fileName)

			// Try copying to ~/.codeforge/pipelines/ if daemon is set up
			home, err := os.UserHomeDir()
			if err == nil {
				destPath := filepath.Join(home, ".codeforge", "pipelines", fileName)
				_ = os.MkdirAll(filepath.Dir(destPath), 0755)
				_ = os.WriteFile(destPath, []byte(kzmContent), 0644)
				color.Green("CodeForge ✓  Config copied to daemon path: %s", destPath)
			}
		} else {
			color.Yellow("Initialization cancelled.")
		}
	},
}

func prompt(r *bufio.Reader, label, defaultValue string) string {
	fmt.Printf("%s [%s]: ", label, defaultValue)
	input, _ := r.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
}

func init() {
	rootCmd.AddCommand(initCmd)
}
