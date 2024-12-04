ALTER TABLE jobs ADD COLUMN search_engine TEXT;

ALTER TABLE jobs ADD COLUMN search_options TEXT;

ALTER TABLE jobs ADD COLUMN query_field TEXT;

ALTER TABLE jobs ADD COLUMN output_field TEXT;

INSERT INTO  "streams" ("name", "description") VALUES ('rag_example', '借助搜索引擎进行RAG');
INSERT INTO "main"."jobs" ("id", "stream_id", "sorted", "name", "type", "description", "llm_model", "temperature", "top_p", "max_tokens", "template", "system_prompt", "method", "endpoint", "params", "output", "output_parses", "created_at", "search_engine", "search_options", "query_field", "output_field") VALUES (1, 1, 1, 'web_search', 'search', '搜索互联网内容', '', NULL, NULL, NULL, '', NULL, '', '', '', '', NULL, '2024-12-04 06:48:46', 'google', '{"block_size":20000,"days":7,"depth_mode":false,"limit":40,"topic":"general"}', 'question', 'search_result');
INSERT INTO "main"."jobs" ("id", "stream_id", "sorted", "name", "type", "description", "llm_model", "temperature", "top_p", "max_tokens", "template", "system_prompt", "method", "endpoint", "params", "output", "output_parses", "created_at", "search_engine", "search_options", "query_field", "output_field") VALUES (2, 1, 2, 'chat_completion', 'prompt', '通过搜索引擎返回的内容回答问题', 'gpt-4o', NULL, NULL, NULL, '问题如下：
{{.question}}

已知信息:
{{.search_result}}

请根据已知信息严格按照要求，并仔细参考示例使用按照输出格式来解读问题：
', '整合已知信息并回答问题，按照以下步骤完成任务：

  1.	整合摘要(look_up)：从提供的已知信息（如新闻文章、维基百科等）中提取与问题主题相关的内容。根据问题的主题，整理和分类信息，生成结构化的总结摘要，覆盖所有重要信息点。。
  2.	推导初步回答(initial_output)：基于整合的摘要，结合问题进行分析与推理，提出一个逻辑清晰的初步回答或总结。这部分应紧扣问题，清晰呈现推导过程。
  3.  得出最终结论(final_output)：将整合的摘要和推导出的初步回答合并，重新列举摘要大纲，明确各部分的论证逻辑：从背景到分析依据，再到结论或答案的合理性。最终的回答段落应清晰流畅，逻辑紧密，重点突出，便于阅读
  3.	语言一致性：无论其他信息的语言为何，最终回答的语言必须与问题所用语言一致。

具体要求：

  1. 结构化输出：每一步骤的内容都需清晰划分，依次呈现：摘要 (look_up)、初步回答 (initial_output)、最终结论 (final_output)。
  2. 准确性与完整性：确保整合的信息全面准确，推理过程符合逻辑，回答覆盖问题的所有关键点。
  3. 语言表达：用语准确，表达简洁；确保语言风格一致且符合问题的语境。

注意事项：

	•	如果已知信息中存在冲突或矛盾，需在推导初步回答时加以辨析，并在最终结论中明确指出。
  •	针对开放性问题，最终答案需考虑多角度分析，避免片面回答。
  •	输出应全面且具体，确保每部分信息充实，无明显遗漏或敷衍。
  •	禁止在回答中使用“根据有关信息”或”已知内容得出“等词语，以便更好地表达推导过程和结论。

输出格式：

以下为标准输出格式，严格遵守该规则在XML标记内生成对应步骤的内容：

  <look_up>
    [在此处列出总结的摘要内容]
  </look_up>

  <initial_output>
    [在此处给出问题的初步回答]
  </initial_output>
  
  <final_output>
    [在此处给出最终答案]
  </final_output>

示例：

  问题：

  哪些企业被在最近因为关税被要求将生产链撤出中国大陆

  输出：

    <look_up>
      1. 美国总统当选人特朗普宣布将对中国进口产品征收额外关税，微软、戴尔和惠普等公司要求中国制造商增加生产以应对即将到来的需求高峰（GIGAZINE, Nikkei Asia）。
      2. 为了规避关税，微软计划将Xbox和Surface的组装迁出中国，戴尔则在东南亚分散生产以降低地缘政治风险，惠普在泰国租赁工厂并计划增加生产（Nikkei Asia）。
      3. 一些电子元件供应商正在加速从中国撤出生产以应对新的关税政策，这一趋势并非直接因特朗普当选，但其政策加速了这一进程（Business Insider）。
      4. Best Buy的CEO表示，特朗普的关税可能导致电子产品价格上涨，并正在与供应商合作以在关税生效前进口更多商品（CBS News）。
      5. 传闻IBM也计划撤出中国，其他外企如苹果和特斯拉也在考虑转移生产基地，原因包括美国政策、成本上升和东南亚国家的招商引资力度（网易新闻）。
    </look_up>

    <initial_output>
      首先，我们从已知信息中了解到，微软、戴尔和惠普等企业因关税原因已要求将生产链撤出中国大陆。
      然后，通过查阅相关报道，我们发现其他除了微软、戴尔、惠普等企业外，可能还有其他企业最近因为关税被要求将生产链撤出中国大陆。
      例如，根据“网易订阅”的报道，苹果、IBM 等高科技外企也有撤离中国的计划。
      此外，Best Buy CEO 在“CBS News”的报道中表示，如果特朗普实施新关税，个人电子产品可能会变得更贵，这也暗示了其他电子企业可能会受到影响。但具体还有哪些企业受到影响，还需要进一步的调查和确认。
    </initial_output>
    
    <final_output>
      除了微软、戴尔和惠普，近期因关税政策而考虑将生产链撤出中国大陆的企业概述：

      近期，美国关税政策对全球科技企业供应链产生了显著影响，促使多家企业重新评估其在中国大陆的生产和供应链布局。以下是一些受到关税政策影响并计划调整生产链的企业：

      1. **IBM**（网易新闻）：
        - **计划**：IBM已宣布关闭其在中国的多个研发部门，并计划将部分业务转移至其他国家。
        - **原因**：美国政策的影响、中国大陆成本上升以及东南亚国家的招商引资力度。

      2. **苹果**（网易新闻）：
        - **考虑**：苹果正在考虑将部分生产迁移出中国，以降低地缘政治风险和关税带来的成本压力。

      3. **特斯拉**（网易新闻）：
        - **考虑**：特斯拉也在评估其供应链策略，以应对潜在的关税影响。

      4. **Best Buy**（CBS News）：
        - **声明**：Best Buy CEO提到，关税可能导致电子产品价格上涨，暗示可能会对其他电子企业产生连锁反应。

      5. **其他电子元件供应商**（Business Insider）：
        - **趋势**：许多电子元件供应商正在加速从中国撤出生产，以应对新的关税政策，尽管这一趋势并非直接因特朗普当选，但其政策加速了这一进程。

      总结来看，除了微软、戴尔和惠普，IBM、苹果、特斯拉等企业正在考虑或已经采取措施，将部分生产链撤出中国大陆。这些举措主要是为了应对美国关税政策的影响，降低地缘政治风险，并利用东南亚国家的招商引资优势。
    </final_output>
', '', '', 'question,search_result', 'look_up,initial_output,final_output', NULL, '2024-12-04 06:48:46', '', NULL, '', '');