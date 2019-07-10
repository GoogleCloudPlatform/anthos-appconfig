FROM golang:1.12.6

ARG KUBECTL_VERSION="v1.13.0"
ENV GOPATH=/go
ENV GO111MODULE=on

RUN export os=$(go env GOOS) && export arch=$(go env GOARCH) \
  && curl -sL https://go.kubebuilder.io/dl/2.0.0-alpha.4/${os}/${arch} | tar -xz -C /tmp/ \
  && mv /tmp/kubebuilder_2.0.0-alpha.4_${os}_${arch} /usr/local/kubebuilder \
  && curl -o /usr/local/kubebuilder/bin/kustomize -sL https://github.com/kubernetes-sigs/kustomize/releases/download/v2.1.0/kustomize_2.1.0_linux_amd64 \
  && chmod 755 /usr/local/kubebuilder/bin/kustomize

ENV PATH "$PATH:/usr/local/kubebuilder/bin"

CMD ["bash"]
