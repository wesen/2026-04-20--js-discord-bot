function greet(name, excited) {
  return {
    greeting: excited ? `Hello, ${name}!` : `Hello, ${name}`,
    mode: excited ? "excited" : "calm"
  };
}

__verb__("greet", {
  short: "Greet one person from the Discord-style bot example",
  fields: {
    name: {
      argument: true,
      help: "Person name"
    },
    excited: {
      type: "bool",
      short: "e",
      help: "Add excitement"
    }
  }
});

function banner(name) {
  return `*** ${name} ***\n`;
}

__verb__("banner", {
  short: "Render a plain-text banner",
  output: "text",
  fields: {
    name: {
      argument: true,
      help: "Banner label"
    }
  }
});
