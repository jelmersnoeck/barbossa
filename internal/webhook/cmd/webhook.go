package cmd

import (
	"crypto/tls"
	"log"

	"github.com/jelmersnoeck/barbossa/internal/kit/signals"
	"github.com/jelmersnoeck/barbossa/internal/webhook"
	"github.com/jelmersnoeck/barbossa/pkg/client/generated/clientset/versioned"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var runFlags struct {
	MasterURL  string
	KubeConfig string

	TLSCertFile string
	TLSKeyFile  string

	Address string
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Run the Webhook Server",
	Run:   runFunc,
}

func runFunc(cmd *cobra.Command, args []string) {
	stopCh := signals.SetupSignalHandler()

	cert, err := tls.LoadX509KeyPair(runFlags.TLSCertFile, runFlags.TLSKeyFile)
	if err != nil {
		log.Fatalf("Filed to load key pair: %v", err)
	}

	cfg, err := clientcmd.BuildConfigFromFlags(runFlags.MasterURL, runFlags.KubeConfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building Kubernetes clientset: %s", err)
	}

	crdClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building Barbossa clientset: %s", err)
	}

	srv, err := webhook.NewServer(kubeClient, crdClient, runFlags.Address, cert)
	if err != nil {
		log.Fatalf("Error configuring the webhook server: %s", err)
	}

	if err := srv.Run(stopCh); err != nil {
		log.Fatalf("Error running the webhook server: %s", err)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringVar(&runFlags.MasterURL, "master-url", "", "The URL of the master API.")
	runCmd.PersistentFlags().StringVar(&runFlags.KubeConfig, "kubeconfig", "", "Kubeconfig which should be used to talk to the API.")

	runCmd.PersistentFlags().StringVar(&runFlags.TLSCertFile, "tls-cert-file", "/certs/tls.crt", "The location for the certificates file.")
	runCmd.PersistentFlags().StringVar(&runFlags.TLSKeyFile, "tls-key-file", "/certs/tls.key", "The location for the certificate private key file.")

	runCmd.PersistentFlags().StringVar(&runFlags.Address, "address", ":6443", "address to run the webhook on")
}
