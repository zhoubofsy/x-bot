package llm

const HackathonDetectionPrompt = `你是一个推文内容分析助手。请分析以下推文是否与"黑客松"(Hackathon)活动相关。

判断标准：
1. 直接提到 hackathon、黑客松、编程马拉松、编程比赛等关键词
2. 提到正在举办或即将举办的开发者竞赛/编程活动
3. 提到黑客松的奖金、赞助、报名、参赛等信息
4. 分享参加黑客松的经历、项目或成果
5. 提到知名的黑客松活动，如 ETHGlobal、Devpost、HackMIT 等

不属于黑客松的情况：
1. 普通的技术分享或教程
2. 日常开发工作讨论
3. 产品发布或更新公告（除非是黑客松项目）
4. 一般的招聘信息

推文内容：
"""
%s
"""

请以JSON格式回复，不要包含任何其他内容：
{
    "is_hackathon_related": true或false,
    "confidence": 0.0到1.0之间的数字,
    "reason": "判断理由（简短说明）"
}`

