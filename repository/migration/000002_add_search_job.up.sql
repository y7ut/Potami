ALTER TABLE jobs ADD COLUMN search_engine TEXT;

ALTER TABLE jobs ADD COLUMN search_options TEXT;

ALTER TABLE jobs ADD COLUMN query_field TEXT;

ALTER TABLE jobs ADD COLUMN output_field TEXT;

INSERT INTO  "streams" ("name", "description") VALUES ('rag_example', '借助搜索引擎进行RAG');
INSERT INTO "main"."jobs" ("id", "stream_id", "sorted", "name", "type", "description", "llm_model", "temperature", "top_p", "max_tokens", "template", "system_prompt", "method", "endpoint", "params", "output", "output_parses", "created_at", "search_engine", "search_options", "query_field", "output_field") VALUES (1, 1, 1, 'web_search', 'search', '搜索互联网内容', '', NULL, NULL, NULL, '', NULL, '', '', '', '', NULL, '2024-12-02 08:00:17', 'google', '{"block_size":6000,"days":7,"depth_mode":true,"limit":30,"topic":"news"}', 'question', 'search_result');
INSERT INTO "main"."jobs" ("id", "stream_id", "sorted", "name", "type", "description", "llm_model", "temperature", "top_p", "max_tokens", "template", "system_prompt", "method", "endpoint", "params", "output", "output_parses", "created_at", "search_engine", "search_options", "query_field", "output_field") VALUES (2, 1, 2, 'chat_completion', 'prompt', '通过搜索引擎返回的内容回答问题', 'gpt-4o', NULL, NULL, NULL, '问题如下：
{{.question}}

已知信息:
{{.search_result}}

在适当的 XML 标记内输出结果：

<look_up>
  [在此处列出摘要总结]
</look_up>

<final_output>
  [在此处插入问题答案思考的结论或答案]
</final_output>
', '仔细阅读问题和已知信息，按照以下步骤和要求并参考示例回答问题：
  1.	整合摘要：将已知信息（包括新闻文章、维基百科等）的每一个段落分别生成清晰、连贯的总结，字数控制在1000字以内
  2.	推导回答：根据摘要内容拟定回答的大纲，并结合问题得到最终的答案，确保答案结构完整且段落清晰，具体段落和结构参考下面的示例。
  3.	语言一致性：无论其他信息的语言为何，最终回答的语言必须与问题所用语言一致。

示例：
  问题：
  台积电最近有哪些重要的司法事件?

  回答：
  台积电近期陷入多项司法及争议事件，涵盖员工歧视诉讼、专利纠纷、扩张政策争议以及国际贸易挑战。这些问题可能对其企业声誉、运营及市场战略造成持续影响。
    1. 美国员工集体诉讼：歧视非亚裔员工（IndustryWeek,Forbes报道）
      •	案件详情：13名前现员工指控台积电存在种族歧视行为，具体表现为：
      •	招聘偏好会中文的候选人，并通过“亚洲猎头”优先招募台湾员工。
      •	非亚裔员工在会议和业务中被孤立，因沟通常以中文进行。
      •	工作环境对非亚裔员工存在敌对性。
      •	案件进展：此案于2024年8月首次提出，后重新提交为集体诉讼。原告来自多个族裔背景，包括美国、墨西哥、尼日利亚、欧洲及韩国。
      •	台积电回应：台积电未直接回应指控细节，仅重申其重视多元化和公平就业。
    2. 与格芯的专利纠纷
      •	争议内容：涉及7nm至28nm制程技术的专利权问题。格芯指控台积电的处理器侵权，这些处理器广泛应用于苹果产品。
      •	潜在影响：虽未直接影响台积电在美扩张计划，但可能对其声誉及与主要客户的合作关系构成威胁。
    3. 美国扩张项目的雇佣政策争议（法新社报道）
      •	背景：台积电在亚利桑那州的工厂建设项目因严重依赖台湾派遣员工而引发争议。
      •	批评声音：过度依赖外籍劳工，被认为不符合美国制造业本地化目标。
      •	与台积电员工歧视诉讼中的指控产生联动效应，加剧公众对其管理文化的关注。
  根据这些事件，我们可以发现台积电正面临多重法律和政策压力，也正是这些挑战突显台积电在全球扩张和应对法律与文化多样性上的深层次矛盾，这些诉讼不仅可能影响其在美国的业务扩展，还可能对其全球市场地位造成影响。
', '', '', 'question,search_result', 'look_up,final_output', NULL, '2024-12-02 08:00:17', '', NULL, '', '');