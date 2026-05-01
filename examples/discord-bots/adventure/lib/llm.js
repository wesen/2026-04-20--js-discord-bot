const llm = require("adventure_llm")
const { parseLLMJson } = require("./schema")

function completeJson(request, onChunk) {
  console.log("[adventure] llm.completeJson request", JSON.stringify({ purpose: request && request.purpose, metadata: request && request.metadata, streaming: Boolean(onChunk && llm.streamJson) }))
  let result = onChunk && llm.streamJson ? llm.streamJson(request, onChunk) : llm.completeJson(request)
  console.log("[adventure] llm.completeJson raw result", JSON.stringify({ ok: result && result.ok, provider: result && result.provider, error: result && result.error, statusCode: result && result.statusCode, usage: result && result.usage, streamed: result && result.streamed }))
  if (onChunk && llm.streamJson && (!result || result.ok !== true)) {
    console.log("[adventure] llm stream failed; falling back to non-stream", JSON.stringify({ error: result && result.error, statusCode: result && result.statusCode }))
    result = llm.completeJson(request)
    console.log("[adventure] llm fallback result", JSON.stringify({ ok: result && result.ok, provider: result && result.provider, error: result && result.error, statusCode: result && result.statusCode, usage: result && result.usage }))
  }
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
