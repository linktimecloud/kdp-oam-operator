# 1. 模块说明
kdp-oam-operator基于KDP大数据模型实现提供高度可编程以应用为核心的交付模式，用户可以基于kdp-oam-operator基本扩展用户自定义服务。在kdp-oam-operator中最关键的是XDefinition、Application、BDC、ContextSetting、ContextSecret，它们都基于K8s CRD定义模型基本结构通过实现独立的controller控制器来完成资源生命周期管理。
![components.png](resources%2Fcomponents.png)
整体来来看kdp-oam-operator的功能模块分为3个部分：基础能力模块、KDP模型controller、用户工具

## 1.1. 基础能力模块
### 1.1.1. template
template是基础能力中最重要的，它为KDP模型提供统一的基础能力底座。由于每个大数据服务拓扑结构不同所需的K8S资源也不同（比如hdfs、kafka会使用statusfulset管理工作负载而flink taskmanager则可以使用depolyment来管理），如果想要通过一个K8S CRD来提供所有大数据服务的抽象支持则必须让这个模型具备一定的可编程能力，将一部分特性的封装交给平台团队的服务构建者。template就是为了处理这个问题，它根据平台团队定义的模板与最终用户提供的安装参数生成服务部署所需的K8S资源。
![template.png](resources%2Ftemplate.png)
kdp-oam-operator默认提供基于cue的template实现

### 1.1.2. parse
parse本质上是服务模型CRD使用template的适配器，通过kdp-oam-operator元数据信息与平台团队定义的服务模版适配调用template能力。
![parser.png](resources%2Fparser.png)

### 1.1.3. schema
为了将平台团队定义的服务部署参数展示给最终用户，需要提供一个统一的数据结构。schema设计上通过json schema描述参数类型与结构、ui schema记录前端展示的效果、error schema声明参数校验异常时的提示。在kdp-oam-operator默认提供OpenAPI json schema与VJSF ui schema，如果用户需要其他类型的schema参数需要额外扩展实现。

如果让平台团队来直接编写schema会加重整个系统的配置维护成本，kdp-oam-operator实现上通过XDefinition parameter自动生成OpenApi JSON schema通过添加的注释信息生成vjsf json schema。
![schema.png](resources%2Fschema.png)


## 1.2. KDP模型controller
### 1.2.1. XDefinition controller
一般K8s operator会通过代码流程将CRD所需生成的K8s资源预编写好，用户创建CR后controller根据预定义好的K8S资源进行组件部署或更新完成组件生命周期管理。但是这种模式无法适配大数据场景庞大的组件生态，每个大数据服务实现一个K8S operator的开发周期太长，而且不能兼容开源通过其他方式(比如helm)提供的服务。XDefinition关键是实现一个高扩展性的可编程模版作为所有KDP模型的底座，为不同大数据服务提供自主封装、灵活配置的能力。

XDefinition中`spec.apuResource`定义资源类型，而`spec.schematic.cue`则是资源模板定义通过cue实现。XDefinition模版CR注册后XDefinition controller会检查cue parameter自动转换成json schema与ui schema并生成ConfigMap进行记录。

其他KDP CRD controller都会将最终用户部署的CR资源与XDefinition模板进行匹配然后生成模板中定义的K8S资源。

### 1.2.2. BDC controller
BDC(bigdata cluster)是通过抽象资源对大数据服务资源隔离与管理。

BDC的具体实现也需要通过XDefinition模版定义，而bdc controller会将最终用户部署的bdc资源与平台团队注册的XDefinition进行匹配然后根据用户定义的参数生成最终的K8s资源。
![bdc-controller.png](resources%2Fbdc-controller.png)

### 1.2.3. Application controller
Application controller是Application CRD对应的控制器，以XDefinition为中心模板根据用户定义的参数生成最终的K8s资源。

为了区分不同大数据服务相较于bdc而言Application多了type定义，需要声明服务的类型。
![application-controller.png](resources%2Fapplication-controller.png)

### 1.2.4. ContextSetting controller与ContextSecret controller
ContextSetting controller是为服务上下文ContextSetting CRD的控制器，ContextSecret controller则是服务敏感信息上下文ContextSecret CRD的控制器。它们都会根据XDefinition模板定义生成K8S资源。
它们的功能逻辑与Application controller的基本逻辑一致这里不再重复展示。