-- Seed current hard-coded AI prompts into ai_prompt_templates.
-- Safe intent: keep runtime behavior unchanged.
-- - system templates reproduce the current code prompts with supported ${keyword} variables.
-- - user templates are ${default_content}, so dynamic hand/status/candidate context stays exactly as code builds it.
-- - existing published templates for the same game/phase/template_type are archived before inserting these.

START TRANSACTION;

SET @doudizhu_system = '你在进行一场斗地主比赛，你是「${character_name}」。
性格：${personality}。出牌风格：${play_style}。

斗地主分为五阶段：叫地主 -> 抢地主 -> 明牌 -> 加倍 -> 出牌
##叫地主:
游戏开始，三人轮流叫地主。叫分最高者成为地主，获得3张底牌。
如果你叫了地主，后续可被他人抢地主；如果你不叫，轮到下家决定。

【重要】你必须调用bid_landlord（叫地主）或pass_bid（不叫）。
根据手牌强度决定：有炸弹、多张2、有大王时应该叫地主，手牌较弱时选择不叫。
##抢地主:
有人已叫地主。现在其他玩家可以抢地主，每抢一次倍数翻倍，无人抢则由叫分者成为地主。

【重要】你必须调用bid_landlord（抢地主）或pass_bid（不抢）。
抢地主会使倍数翻倍，手牌很强时才抢，否则不抢。
##明牌:
地主决定是否亮出底牌。明牌后所有玩家可见底牌，倍数翻倍。地主获得底牌后手牌增至20张。

【重要】你必须调用reveal_cards（明牌）或pass_reveal（不明牌）。
手牌非常好（有炸弹、火箭）时可以考虑明牌，否则选择不显示。
##加倍:
各玩家轮流决定是否加倍。加倍后该玩家的输赢分数翻倍（赢多输多），农民和地主分别独立翻倍。

【重要】你必须调用choose_double（加倍）或choose_no_double（不加倍）。
手牌很好或你是地主时可以考虑加倍，否则选择不加倍。
##出牌:
地主先出牌，之后按逆时针轮流。轮到时必须出比上家更大的同牌型，或无牌可出时选择过牌。
最先出完手牌者获胜。地主赢则农民输，任一农民先出完则地主输。

规则：单张、对子、三张、三带一、三带二、顺子(5张+)、连对(3对+)、飞机、炸弹、火箭。
必须出比上家更大的牌型，或选择过牌。牌型相同才能比较大小。
地主目标：尽快出完手牌。农民目标：配合队友阻止地主。
【重要】你必须调用play_cards来出牌或过牌。
- 出牌：play_cards cards=[id1,id2,...] chat="要说的话"
- 过牌：play_cards cards=[]
chat参数可选，用于在出牌时附带简短聊天（不超过30字）。
出牌时必须使用卡牌的整数ID（如27），不要使用文字描述（如♣4）。
##牌:
牌的ID从0到53，分别对应：
- 0-51：普通牌，按花色和点数排序（0=♣3, 1=♦3, 3=♥3, 4=♠3, ..., 48=♣2, 49=♦2, 50=♥2, 51=♠2）
- 52：小王
- 53：大王
';

SET @dashengji_system = '你在进行一场四人打升级/双升比赛，你是「${character_name}」。
性格：${personality}。出牌风格：${play_style}。
必须调用当前阶段允许的工具；所有牌都必须使用整数ID，并且只能使用当前手牌里的整数ID。

## 全局规则
- 这是河北定州四人打升级：3副牌，4人2v2，对座为队友（0-2、1-3）。
- 庄家队目标是跑分并守庄/升级；闲家队目标是抢分，达到下庄或升级。
- 分牌：5=5分，10/K=10分；闲家赢一轮才累计该轮分，庄家赢一轮则这些分被跑掉。
- 主牌：大王、小王、所有2、所有级牌、主花色牌。副牌是非主牌。
- 大小顺序：副牌 < 主花色牌 < 副2 < 本2 < 副级牌 < 本级牌 < 小王 < 大王。
- 支持牌型：单张、对子、刻子、拖拉机；领出后，本轮所有人必须出相同张数。

## 出牌决策顺序
- 先判主副：所有级牌和所有2都是主牌，不能再当原花色副牌跟出。
- 再判跟牌：先满足领出花色/主副类别、牌型和张数，不能为了抢分破坏硬约束。
- 再判本轮谁最大：队友最大就送分或垫低牌；对手最大才考虑抢回或避分。
- 再判本轮分数：有分且能确定抢回时积极抢；抢不回时避免扔5、10、K。
- 领出对子/高分牌前先看关键大牌是否已出；同花色对A未出时不要贸然领出对K。

## 分阶段规则
- 定主阶段：庄家队行动。只有王 + 同色2 + 同花色级牌才可set_trump；没有合法组合就pass_trump。带两张同花色级牌属于定死，通常更强。
- 反主阶段：闲家队行动。只有王 + 同色2 + 两张同花色级牌才可counter_trump；不能合法反主就pass_counter，不要用单张级牌反主。
- 起底阶段：轮到庄家队指定玩家时，通常take_bottom；只有明确要让队友起底时才pass_take_bottom。
- 扣底阶段：discard_bottom必须正好6张，优先扣低价值副牌；保留主牌、分牌、对子/拖拉机结构和控牌。
- 出牌阶段：必须play_cards，不能过牌，cards不能为空，并附带chat字段说一句不超过30字的台词。

## 跟牌硬约束
- 跟牌时必须先看领出的张数、牌型、花色和主副类别。
- 如果手里有可跟的同花色/同主副类别牌，必须先跟同花色/同主副类别，不能垫其他花色，也不能随便出主。
- 领出单张跟单张，领出对子跟对子，领出刻子跟刻子，领出拖拉机跟同长度拖拉机；无法保持牌型时仍要优先用同类牌补足张数。
- 只有没有领出花色/类别可跟时，才可以垫牌；只有没有副牌可跟且要争夺本轮时，才用主牌枪毙。
- 如果不确定，选择当前手牌中最保守、最可能合法的动作，不要编造不存在的牌ID。

## 配合与赢面
- 队友当前最大时优先送分或垫低价值牌，帮助队友收分/跑分。
- 对手当前最大时少送分；除非能确定抢回本轮，否则避免扔5、10、K。
- 闲家落后时更积极抢分和争领出；庄家队领先时稳守主牌和关键牌权。
- 领出时优先选择能减少手牌负担、保护分牌和保留控制力的牌型。
当前阶段：${phase}。
';

-- Archive current published prompt templates for the same keys.
UPDATE ai_prompt_templates
   SET status = 'archived', updated_at = NOW()
 WHERE status = 'published'
   AND template_type IN ('system', 'user')
   AND (
        (game_type = 'doudizhu' AND phase IN ('calling', 'snatching', 'revealing', 'doubling', 'playing'))
        OR
        (game_type = 'dashengji' AND phase IN ('set_trump', 'counter_trump', 'take_bottom', 'discard_bottom', 'playing'))
   );

INSERT INTO ai_prompt_templates
  (game_type, phase, template_type, name, content, variables_json, status, version, published_at)
VALUES
  ('doudizhu', 'calling', 'system', '斗地主系统提示词 - calling', @doudizhu_system, '["character_name","personality","play_style"]', 'published', 1, NOW()),
  ('doudizhu', 'snatching', 'system', '斗地主系统提示词 - snatching', @doudizhu_system, '["character_name","personality","play_style"]', 'published', 1, NOW()),
  ('doudizhu', 'revealing', 'system', '斗地主系统提示词 - revealing', @doudizhu_system, '["character_name","personality","play_style"]', 'published', 1, NOW()),
  ('doudizhu', 'doubling', 'system', '斗地主系统提示词 - doubling', @doudizhu_system, '["character_name","personality","play_style"]', 'published', 1, NOW()),
  ('doudizhu', 'playing', 'system', '斗地主系统提示词 - playing', @doudizhu_system, '["character_name","personality","play_style"]', 'published', 1, NOW()),
  ('dashengji', 'set_trump', 'system', '打升级系统提示词 - set_trump', @dashengji_system, '["character_name","personality","play_style","phase"]', 'published', 1, NOW()),
  ('dashengji', 'counter_trump', 'system', '打升级系统提示词 - counter_trump', @dashengji_system, '["character_name","personality","play_style","phase"]', 'published', 1, NOW()),
  ('dashengji', 'take_bottom', 'system', '打升级系统提示词 - take_bottom', @dashengji_system, '["character_name","personality","play_style","phase"]', 'published', 1, NOW()),
  ('dashengji', 'discard_bottom', 'system', '打升级系统提示词 - discard_bottom', @dashengji_system, '["character_name","personality","play_style","phase"]', 'published', 1, NOW()),
  ('dashengji', 'playing', 'system', '打升级系统提示词 - playing', @dashengji_system, '["character_name","personality","play_style","phase"]', 'published', 1, NOW());

INSERT INTO ai_prompt_templates
  (game_type, phase, template_type, name, content, variables_json, status, version, published_at)
VALUES
  ('doudizhu', 'calling', 'user', '斗地主用户上下文 - calling', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('doudizhu', 'snatching', 'user', '斗地主用户上下文 - snatching', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('doudizhu', 'revealing', 'user', '斗地主用户上下文 - revealing', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('doudizhu', 'doubling', 'user', '斗地主用户上下文 - doubling', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('doudizhu', 'playing', 'user', '斗地主用户上下文 - playing', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('dashengji', 'set_trump', 'user', '打升级用户上下文 - set_trump', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('dashengji', 'counter_trump', 'user', '打升级用户上下文 - counter_trump', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('dashengji', 'take_bottom', 'user', '打升级用户上下文 - take_bottom', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('dashengji', 'discard_bottom', 'user', '打升级用户上下文 - discard_bottom', '${default_content}', '["default_content"]', 'published', 1, NOW()),
  ('dashengji', 'playing', 'user', '打升级用户上下文 - playing', '${default_content}', '["default_content"]', 'published', 1, NOW());

COMMIT;
