/*
Copyright 2019 The Knative Authors

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

package utils

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"go.uber.org/zap"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
)

// KlogVerbosityKey is the key in config-logging ConfigMap for klog verbosity level.
const KlogVerbosityKey = "klog-verbosity"

// GetLoggingConfig fetches the logging ConfigMap from the given namespace and
// parses it into a *logging.Config. If the ConfigMap is not found, it returns
// the default logging config.
func GetLoggingConfig(ctx context.Context, namespace, loggingConfigMapName string) (*logging.Config, error) {
	loggingConfigMap, err := kubeclient.Get(ctx).CoreV1().ConfigMaps(namespace).Get(ctx, loggingConfigMapName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return logging.NewConfigFromMap(nil)
	} else if err != nil {
		return nil, err
	}
	return logging.NewConfigFromConfigMap(loggingConfigMap)
}

// klogFlagSet is initialized once at package load; reused on every verbosity update.
// klogMu guards all writes to klogFlagSet — flag.FlagSet.Set is not goroutine-safe.
var (
	klogFlagSet = func() *flag.FlagSet {
		fs := flag.NewFlagSet("klog", flag.ContinueOnError)
		klog.InitFlags(fs)
		return fs
	}()
	klogMu sync.Mutex
)

// SetKlogVerbosityFromConfigMap reads klog-verbosity from the ConfigMap data and
// applies it to klog. Missing or empty values are no-ops; "0" resets verbosity.
// Valid range is 0–9.
func SetKlogVerbosityFromConfigMap(data map[string]string) error {
	level, ok := data[KlogVerbosityKey]
	if !ok || level == "" {
		return nil
	}

	n, err := strconv.Atoi(level)
	if err != nil || n < 0 || n > 9 {
		return fmt.Errorf("invalid %s value %q: must be an integer between 0 and 9", KlogVerbosityKey, level)
	}

	klogMu.Lock()
	defer klogMu.Unlock()
	return klogFlagSet.Set("v", level)
}

// UpdateKlogVerbosityFromConfigMap returns a ConfigMap watch handler that updates
// klog verbosity when the config-logging ConfigMap changes.
func UpdateKlogVerbosityFromConfigMap(logger *zap.SugaredLogger) func(*corev1.ConfigMap) {
	return func(cm *corev1.ConfigMap) {
		level := cm.Data[KlogVerbosityKey]
		if err := SetKlogVerbosityFromConfigMap(cm.Data); err != nil {
			logger.Warnw("Failed to update klog verbosity", zap.Error(err))
			return
		}
		if level != "" && level != "0" {
			logger.Infow("Updated klog verbosity", zap.String("level", level))
		}
	}
}
