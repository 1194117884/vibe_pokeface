import { describe, expect, it } from "vitest";
import { formatDashengjiNoticeToast, mergeDashengjiNoticePlayers } from "../dashengji-notice";

const players = [
  { seat: 0, nickname: "yongkl" },
  { seat: 1, nickname: "妈妈" },
  { seat: 2, nickname: "张总" },
  { seat: 3, nickname: "老李" },
];

describe("formatDashengjiNoticeToast", () => {
  it("formats set trump with the trump suit result", () => {
    expect(
      formatDashengjiNoticeToast(
        { seq: 1, kind: "trump_decision", seat: 0, action: "set_trump" },
        players,
        2,
      ),
    ).toBe("yongkl:定主-梅花");
  });

  it("formats pass trump without a result suffix", () => {
    expect(
      formatDashengjiNoticeToast(
        { seq: 1, kind: "trump_decision", seat: 3, action: "pass_trump" },
        players,
        2,
      ),
    ).toBe("老李:不定主");
  });

  it("formats counter trump with the trump suit result", () => {
    expect(
      formatDashengjiNoticeToast(
        { seq: 1, kind: "counter_decision", seat: 1, action: "counter_trump" },
        players,
        1,
      ),
    ).toBe("妈妈:反主-红桃");
  });

  it("formats pass counter without a result suffix", () => {
    expect(
      formatDashengjiNoticeToast(
        { seq: 1, kind: "counter_decision", seat: 2, action: "pass_counter" },
        players,
        1,
      ),
    ).toBe("张总:不反主");
  });

  it("formats take bottom and discard bottom notices", () => {
    expect(
      formatDashengjiNoticeToast(
        { seq: 1, kind: "bottom_taken", seat: 4 },
        players,
        1,
      ),
    ).toBe("玩家5:起底了");
    expect(
      formatDashengjiNoticeToast(
        { seq: 2, kind: "bottom_discarded", seat: 0 },
        players,
        1,
      ),
    ).toBe("yongkl:完成扣底");
  });

  it("uses cached names when state update players omit nicknames", () => {
    const merged = mergeDashengjiNoticePlayers(
      [{ seat: 0 }, { seat: 1, nickname: "新名字" }],
      players,
    );

    expect(
      formatDashengjiNoticeToast(
        { seq: 1, kind: "trick_winner", seat: 0 },
        merged,
        1,
      ),
    ).toBe("yongkl 牌最大");
    expect(
      formatDashengjiNoticeToast(
        { seq: 2, kind: "trick_winner", seat: 1 },
        merged,
        1,
      ),
    ).toBe("新名字 牌最大");
  });
});
