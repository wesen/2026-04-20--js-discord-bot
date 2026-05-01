const llm = require("adventure_llm")
const { parseLLMJson } = require("./schema")

function completeJson(request) {
  const result = llm.completeJson(request)
  if (!result || result.ok !== true) {
    return {
      ok: false,
      error: (result && result.error) || "LLM request failed",
      retryable: Boolean(result && result.retryable),
      raw: result || null,
    }
  }
  const parsed = parseLLMJson(result.text)
  if (!parsed.ok) {
    return { ok: false, error: parsed.error, rawText: result.text, raw: result, parsed }
  }
  return { ok: true, value: parsed.value, rawText: result.text, raw: result, parsed }
}

module.exports = { completeJson }
