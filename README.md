# Potami 波塔米

Potami 是一个构建和发布 AI 驱动型应用的框架，通过核心任务流的高效调度和灵活的数据流转。提供全面的应用配置管理、应用发布、优先级控制、实时监控等特性，使用户可以便捷地管理和创建复杂的AI应用。Potami的设计目标是简化AI驱动型应用开发和管理成本，确保高并发处理的稳定性的同时，也具备高度自定义的可扩展和维护性。

## 核心特性

  1. 并发控制与优先级控制：
  通过协程池，Potami能够有效控制系统的并发执行量，确保在高负载情况下保持系统的平稳运行。用户还可以为任务流（AI应用）设置优先级，保障高优先级任务在资源紧张时优先执行。
  2. 任务流的动态更新：
  Potami允许用户在服务运行期间无需重启即可更新AI应用的任务流的配置。用户可以根据业务需求，随时调整任务的执行步骤和节点设置，极大提高了系统的灵活性和适应性。
  3. 管道化数据流转：
  Potami的任务流内部基于管道式的数据流结构，支持输入和输出的灵活定义和流转。每个任务节点可以通过prompt节点与LLM进行交互，或通过tool节点请求外部API。prompt节点还支持多阶段文本生成任务（如翻译、大纲整理、标题生成等），提供输入和输出的数据结构，满足复杂文本生成任务的需求。
  4. 实时监控与反馈：
  • SSE状态推送：支持通过服务器推送事件 (SSE) 实时获取任务流状态，用户可以随时查看任务的进展情况。
  • 任务进度查询：Potami将每个任务的详细进度信息存储在Redis中，用户可随时查询当前任务的状态。
  • Callback结果回调：通过异步Callback的方式，任务完成后自动回调结果，适用于需要即时反馈的场景。
  5. CLI客户端支持：
  Potami提供了命令行客户端 (CLI) 工具，用于任务流的创建、管理与调试。通过CLI，用户可以快速管理任务流配置、执行流调试，并监控任务状态，极大简化了操作流程。

Potami系统结合了强大的任务流控制、实时监控与多种反馈机制，让用户能够高效地使用管理复杂的任务流配置，从而更充分地利用大语言模型的能力。

## Stream 任务流

任务流是 Potami 系统的核心也是一个AI应用的关键部分，它可以视为由多个 LLM 和外部工具（Tool）节点构成的操作流程。借助于 Stream 的功能，我们可以高效地管理复杂的 Prompt Engineering 流程。

每个 Job 的主要工作是根据 Params 中的输入参数执行任务，并生成 Outputs。不同类型的 Job 的 Output 形式可能有所不同，但它们都遵循以下几条基本规则。这些规则可以帮助我们轻松地配置和创建一个有效的 Stream：

  1. 数据流转规则：
  在整个 Stream 流程中，Params 和 Outputs 的数据会自上而下逐步流转。流程中后续任务所需的 Params 必须是前面任务所提供的数据或其 Outputs 中的内容。
  2. 数据覆盖：
  当 Params 和 Outputs 中包含相同的参数时，每个阶段执行后，这些参数的值会被更新，并且在接下来的流程中可以再次覆盖，确保数据流的一致性。

通过遵循这些规则，您可以创建一个自定义的 Stream，实现复杂的操作流程并高效管理数据的流转和处理。

## Jobs 任务节点

一个 Stream 由一系列内部任务（Jobs）组成，主要包括以下几种任务类型：

### Prompt Dialog

提示词对话文本生成任务节点，基础 Prompt 类型，用于与 LLM 交互。

对于 prompt 类型的 Job，其 Params 和 Outputs 中的参数必须在 Prompt template 或 System Prompt 中体现，以确保 LLM 任务能够正确接收和使用这些参数。
  
### Search

联网搜索任务节点，可以通过搜索的API使模型获得联网能力，目前支持的搜索引擎有`Tavily`和`Google Custom Search`。

对于 search类型的节点，必须提供 QueryField 和 OutputField两个属性参数，用来标记输入和输出。

### API TOOL

API调用任务节点，用于通过 HTTP 接口调用外部API，扩展和交互数据处理的能力。

对于 api_tool 类型的 Job，需要使用 OutputParse 和 jsonpath 来提取外部 API 响应中的所需数据并构建 Outputs。

## Schema 配置描述文件

常规的任务流需要指定一个名称，并且由最少一个阶段的`Job`构成，下面就是一个单阶段的`Stream`的示例, 里面只有一个`Job`节点。

```yaml
name: "article"
description: "文章创作-单阶段"
jobs:
  - name: "article_write"
    type: "prompt"
    description: "多阶段文本创作"
    llm_model: "gpt-4o"
    temperature: 1
    system_prompt: |
      你是一名高水平的文字编辑，负责根据各种类型的企业事件文章，创作原创的新闻文章。请仔细按照以下步骤说明完成创作任务：
      步骤一：文本内容转换
      当文本内容不是中文，需要按照原文内容详细逐步翻译，并在接下来的创作时使用翻译后的内容来写作，在翻译时需要注意下列事宜：
      (i) 准确性（通过纠正添加、误译、遗漏或未翻译文本的错误）。
      (ii) 流畅性（通过应用中文语法、拼写和标点符号规则，并确保没有不必要的重复）。
      (iii) 风格（通过确保翻译反映源文本的风格并考虑任何文化背景）。
      (iv) 术语（通过确保术语使用一致并反映源文本领域；并且仅确保使用等效的中文习语）。
      步骤二：大纲整理
      提取文本内容的文本内容中所描述的重要事件（例如技术进展，人事变动，企业事件等），建立大纲，用于在创作时参考，文章大纲需要采用下列方法：
      (i) 理解内容：全面阅读新闻稿，理解其主要内容、背景和意义。
      (ii) 提取信息：列出关键事实、人物、事件和时间。
      (iii)构思新角度：思考如何从不同的视角来重述这些信息。
      步骤三：文章写作
      当一切准备完成后，你要准守下面的写作方式来创作：
      文章主要分为，标题，概括，正文三部分。
      (i)文章的标题需要参考文章大纲中的重要事件，交代主要事件内容和背景或时间
      (ii)文章的概括则要突出内容的重点，用简短的篇幅来概述，不超过40字
      (iii)正文内容采用三段式，根据原文内容或者翻译后的内容来创作，并请遵循下列的规则来创作：
      第一段， 用来介绍文本内容所主要内容、背景和意义
      第二段， 根据文章大纲描述的事件始末，最新进展，列出关键事件，人物和时间
      第三段， 从不同的视角来重述这些信息，介绍事件后续
      输出：
      根据要求，在对应的 XML 标签内输出过程中的每个步骤的结果：
      <translation>
      [如果是非中文内容，在此插入翻译]
      </translation>
      <outline>
      [在此处插入整理后的文章的大纲]
      </outline>
      <title>
      [在此插入为文章创作的标题]
      </title>
      <overview>
      [在此插入内容摘要]
      </overview>
      <content>
      [在此插入文章的正文段落]
      </content>

    template: |
      请根据下方文本内容来进行创作:
      {{.text}}

    params:
      - "text"

    output:
      - "translation"
      - "outline"
      - "title"
      - "overview"
      - "content"

```

要使用这个Stream ,只需要提供一个文本属性的内容 text 即可。

#### 一些其他类型Job的示例

必要时还可以添加一个Job用来生成文章的摘要

```yaml
  - name: overview_generator
    type: prompt
    description: 摘要生成
    llm_model: gpt-4o
    temperature: 0.7
    system_prompt: |
      你是一名有多年经验的文字内容创作者, 擅长在团队中的文案写作和稿件创作工作。
      本次的工作内容是根据创作后的文章正文内容根据已经确定的选题来制定稿件的标签和副标题。
      标签最多可以包含3个，每个最多4个字，标签之间用#隔开，标签中不能包含空格（例如： #标签1 #标签2 #标签3）。
      副标题体现文章核心摘要，采用头条新闻的风格，能够更好的引起读者的注意力。
      接下来会有其他编辑根据你制定的标签和副标题，来进行下一步的稿件撰写工作。所以请严格按照工作规范来完成，这将会影响你的职业生涯。
      
      输出：
      根据要求，在对应的 XML 标签内输出过程中的每个步骤的结果：

      <tags>
      [在此处插入标签]
      </tags>

      <subtitle>
      [在此处插入副标题]
      </subtitle>
    template: |
      文章的选题为《{{.overview}}》请根据下方文本内容来制定标签和副标题:  
      {{.content}}
    params:
      - "overview"
      - "content"
    output:
      - "tags"
      - "subtitle"
```

联网搜索相关信息

```yaml
  - name: search_news
    type: search
    description: 搜索新闻
    search_engine: tavily
    search_options:
      topic: news
      days: 7
      limit: 3
      block_size: 6000
      depth_mode: true
    query_field: tags
    output_field: search_result
```

去检索一个图片用来当封面图，

```yaml

  - name: cover_search
    type: api_tool
    description: 搜索封面
    endpoint: https://picsearch.y7ut.com/v1/pic/search
    method: GET
    params:
      - "tags"
    output_parses:
      cover: "$.data.[0].url"

```

通过api将文稿加入草稿箱

```yaml
# ...
# 追加Jobs


  - name: save_draftbox
    type: api_tool
    description: 保存草稿箱
    endpoint: https://new.y7ut.com/v1/draft/save
    method: POST
    params:
      - "title"
      - "subtitle"
      - "tags"
      - "overview"
      - "content"
      - "cover"
    output_parses:
      news_id: "$.data.news_id"

```

## potactl

### 上下文管理

```sh
potactl context ls
```

可以通过切换上下文来使用不同环境的`Potami`服务，其他的子命令可以使用`-c`来直接切换本次的上下文
默认的上下文为本地的6180端口，即local `http://127.0.0.1:6180`

可以使用 set 命令指定默认的上下文环境

```sh
potactl context set local
```

看见下面输出时，设定成功

```sh
current context is local
endpoint is http://127.0.0.1:6180
```

### 创建Stream

```sh
potactl stream apply -f config/stream_example/company_qa_baidu.yaml
```

当然也可以从标准输出传入yaml文本内容, 或直接通过参数传入, 

### 查看全部Stream

```sh
potactl stream list 
```

### 查看Stream详细信息

```sh
potactl stream info company_qa
```

可以使用`-o --output`来查看`json`或`yaml`格式详情

```sh
potactl stream info company_qa -o json
```

### 发布 Stream

发布任务流可以通过API的形式也可以使用`potactl`，更推荐使用potactl来进行

```sh
cat stream.yaml | potactl stream apply
```

或者

```sh
potactl stream apply -f stream.yaml
```

### 使用 Stream

同样我们可以使用`potactl`来进行调试， 使用`-p`来携带属性和值作为`Params`，发起指定的stream任务

```sh
potactl stream complete company_qa_slim -p company="长沙景嘉微电子股份有限公司" -p question="企业有什么主要产品"
```

正式使用时，往往会通过API方式来进行调用，这时`Potami`提供三种方式来进行使用分别为：

1. 同步 sync
2. 异步 async
3. 流式 stream

具体见接口文档部分，这三种方式也都分别支持`CallBack`（执行任务完成后，将结果或失败的详情，根据提供的地址发器回调）和进度查询（更新Redis中的表示进度的KEY）


## API

施工中🏗️...
