/*
Copyright 2021 The Clusternet Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"fmt"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/clusternet/clusternet/pkg/features"
	"github.com/clusternet/clusternet/pkg/known"
	"github.com/clusternet/clusternet/pkg/utils"
)

// AgentOptions holds the command-line options for command
type AgentOptions struct {
	*ClusterRegistrationOptions
	*utils.ControllerOptions

	// PredictorAddress specifies the address of predictor
	PredictorAddress string
	// PredictorDirectAccess indicates whether the predictor can be accessed directly by clusternet-scheduler
	PredictorDirectAccess bool
	// PredictorPort specifies the port on which to serve built-in predictor
	PredictorPort int
	// ServeInternalPredictor indicates whether to serve built-in predictor. It is not a flag.
	ServeInternalPredictor bool

	// No tunnel logging by default
	TunnelLogging bool

	// Flags hold the parsed CLI flags.
	Flags *cliflag.NamedFlagSets
}

// Complete completes all the required options.
func (opts *AgentOptions) Complete() error {
	allErrs := []error{}

	// complete cluster registration options
	errs := opts.ClusterRegistrationOptions.Complete()
	allErrs = append(allErrs, errs...)

	// complete leader election and client connection options
	if err := opts.ControllerOptions.Complete(); err != nil {
		allErrs = append(allErrs, err)
	}

	if utilfeature.DefaultFeatureGate.Enabled(features.Predictor) && opts.PredictorAddress == "" {
		opts.PredictorAddress = fmt.Sprintf("http://localhost:%d", opts.PredictorPort)
		opts.ServeInternalPredictor = true
	}

	return utilerrors.NewAggregate(allErrs)
}

// Validate validates all the required options.
func (opts *AgentOptions) Validate() error {
	allErrs := []error{}

	// validate cluster registration options
	errs := opts.ClusterRegistrationOptions.Validate()
	allErrs = append(allErrs, errs...)

	// validate leader election and client connection options
	if err := opts.ControllerOptions.Validate(); err != nil {
		allErrs = append(allErrs, err)
	}

	return utilerrors.NewAggregate(allErrs)
}

// initFlags initializes flags by section name.
func (opts *AgentOptions) initFlags() {
	if opts.Flags != nil {
		return
	}

	fss := &cliflag.NamedFlagSets{}
	// flags for cluster registration
	opts.ClusterRegistrationOptions.AddFlagSets(fss)
	// flags for leader election and client connection
	opts.ControllerOptions.AddFlagSets(fss)

	predictorfs := fss.FlagSet("cluster capacity predictor")
	predictorfs.StringVar(&opts.PredictorAddress, PredictorAddress, opts.PredictorAddress,
		"Set address of external predictor, such as https://abc.com:8080. If not set, built-in predictor will be used when feature gate 'Predictor' is enabled.")
	predictorfs.BoolVar(&opts.PredictorDirectAccess, PredictorDirectAccess, opts.PredictorDirectAccess,
		"Whether the predictor be accessed directly by clusternet-scheduler")
	predictorfs.IntVar(&opts.PredictorPort, PredictorPort, opts.PredictorPort,
		"Set port on which to serve built-in predictor server. It is only used when feature gate 'Predictor' is enabled and '--predictor-addr' is not set.")

	utilfeature.DefaultMutableFeatureGate.AddFlag(fss.FlagSet("feature gate"))

	misc := fss.FlagSet("misc")
	misc.BoolVar(&opts.TunnelLogging, "enable-tunnel-logging", opts.TunnelLogging, "Enable tunnel logging")

	opts.Flags = fss
}

// NewOptions creates a new *options with sane defaults
func NewOptions() (*AgentOptions, error) {
	controllerOptions, err := utils.NewControllerOptions("clusternet-agent", known.ClusternetSystemNamespace)
	if err != nil {
		return nil, err
	}

	opts := &AgentOptions{
		ClusterRegistrationOptions: NewClusterRegistrationOptions(),
		ControllerOptions:          controllerOptions,
		PredictorPort:              8080,
		PredictorDirectAccess:      false,
	}
	opts.initFlags()
	return opts, nil
}
