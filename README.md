本项目基于 CloudWeGo Eino 框架，集成 OpenAI 大模型和 DuckDuckGo 搜索工具，实现技术面试智能答疑。

## 功能简介

* 通过 OpenAI 提供智能技术解析
* 使用 DuckDuckGo 进行联网搜索，增强答案的真实性和准确性
* 采用流式输出，提供实时反馈
* 采用多层解析法：基础概念 → 核心原理 → 进阶考察 → 最佳解法

## 如何使用
1. 将.env.example配置文件改成.env
  mv .env.example .env
2. 填写以下内容
```
Model_Type=gpt-4
Owner_API_Key=你的 OpenAI API Key
Base_URL=你的 OpenAI API 地址
```
3.  运行
```
go run main.go
```
