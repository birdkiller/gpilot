package llm

const SystemPromptRootCause = `你是一位资深的 Kubernetes/云原生运维专家。你的任务是分析告警信息并给出根因分析。

请按以下格式输出你的分析结果：

## 根因分析
[1-3句话描述根本原因]

## 影响范围
[评估此告警的影响范围和严重程度]

## 建议严重级别
[critical/warning/info 中选择一个]

## 修复建议
[按优先级排列的修复步骤，每步包含标题和具体操作]

请基于提供的上下文信息进行分析，确保建议具有可操作性。`

const SystemPromptLogAnalysis = `你是一位资深的 Kubernetes/云原生运维专家。你的任务是分析日志内容，识别错误模式并给出解决建议。

请按以下格式输出：

## 错误摘要
[概括日志中发现的主要问题]

## 错误模式
[识别出的错误模式及其频率]

## 根因推断
[基于日志内容推断的可能根因]

## 修复建议
[具体的修复步骤]

请确保分析基于实际日志内容，避免猜测。`

const SystemPromptNLToLogQL = `你是一位 LogQL 查询专家。用户会用自然语言描述他们想要查询的日志，你需要将其转换为有效的 LogQL 查询语句。

规则：
1. 输出格式必须是有效的 LogQL
2. 使用常见的 label 过滤：namespace, pod, container, app
3. 使用管道操作符进行内容过滤：|=, |~, !=
4. 如果需要统计，使用聚合函数

请按以下格式输出：

## LogQL查询
` + "```" + `
[LogQL查询语句]
` + "```" + `

## 查询说明
[解释这个查询的含义]`

const SystemPromptEventDiagnosis = `你是一位资深的 Kubernetes 运维专家。请分析以下 K8s 事件，给出通俗易懂的诊断说明和处理建议。

请按以下格式输出：

## 事件解读
[用通俗的语言解释这个事件意味着什么]

## 可能原因
[列举可能导致此事件的原因]

## 处理建议
[具体的处理步骤]`

// BuildRootCausePrompt constructs the user prompt for root cause analysis
func BuildRootCausePrompt(alertName, severity, description string, labels map[string]string,
	relatedAlerts []string, recentLogs []string, k8sEvents []string) string {

	prompt := "## 告警信息\n"
	prompt += "- 名称: " + alertName + "\n"
	prompt += "- 严重级别: " + severity + "\n"
	prompt += "- 描述: " + description + "\n"
	prompt += "- 标签:\n"
	for k, v := range labels {
		prompt += "  - " + k + ": " + v + "\n"
	}

	if len(relatedAlerts) > 0 {
		prompt += "\n## 关联告警\n"
		for _, a := range relatedAlerts {
			prompt += "- " + a + "\n"
		}
	}

	if len(recentLogs) > 0 {
		prompt += "\n## 最近日志 (最近15分钟)\n"
		for _, l := range recentLogs {
			prompt += l + "\n"
		}
	}

	if len(k8sEvents) > 0 {
		prompt += "\n## K8s 事件\n"
		for _, e := range k8sEvents {
			prompt += "- " + e + "\n"
		}
	}

	return prompt
}

// BuildLogAnalysisPrompt constructs the prompt for log analysis
func BuildLogAnalysisPrompt(logs []string, contextInfo string) string {
	prompt := "## 日志内容\n```\n"
	for _, l := range logs {
		prompt += l + "\n"
	}
	prompt += "```\n"

	if contextInfo != "" {
		prompt += "\n## 上下文信息\n" + contextInfo + "\n"
	}

	return prompt
}

// BuildNLQueryPrompt constructs the prompt for natural language to LogQL translation
func BuildNLQueryPrompt(question string) string {
	return "请将以下自然语言查询转换为 LogQL 查询语句：\n\n" + question
}

// BuildEventDiagnosisPrompt constructs the prompt for K8s event diagnosis
func BuildEventDiagnosisPrompt(eventType, reason, message, namespace, objectKind, objectName string) string {
	prompt := "## K8s 事件详情\n"
	prompt += "- 类型: " + eventType + "\n"
	prompt += "- 原因: " + reason + "\n"
	prompt += "- 消息: " + message + "\n"
	prompt += "- 命名空间: " + namespace + "\n"
	prompt += "- 相关对象: " + objectKind + "/" + objectName + "\n"
	return prompt
}
