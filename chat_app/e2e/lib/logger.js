/**
 * Logger utility for formatted test output
 */

const colors = {
  reset: "\x1b[0m",
  bright: "\x1b[1m",
  green: "\x1b[32m",
  red: "\x1b[31m",
  yellow: "\x1b[33m",
  blue: "\x1b[34m",
  cyan: "\x1b[36m",
  gray: "\x1b[90m"
};

function header(title) {
  console.log("\n" + colors.bright + colors.cyan + "=".repeat(60) + colors.reset);
  console.log(colors.bright + colors.cyan + `  ${title}` + colors.reset);
  console.log(colors.bright + colors.cyan + "=".repeat(60) + colors.reset);
}

function subheader(title) {
  console.log("\n" + colors.bright + colors.yellow + `--- ${title} ---` + colors.reset);
}

function pass(testName, details = "") {
  const detailStr = details ? colors.gray + ` (${details})` + colors.reset : "";
  console.log(`  ${colors.green}✔ PASS${colors.reset} ${testName}${detailStr}`);
}

function fail(testName, error) {
  console.log(`  ${colors.red}✖ FAIL${colors.reset} ${testName}`);
  if (error) {
    const errorMsg = error.stack || error.message || String(error);
    console.log(colors.red + `     Error: ${errorMsg}` + colors.reset);
  }
}

function info(message) {
  console.log(`  ${colors.blue}ℹ${colors.reset} ${message}`);
}

function warn(message) {
  console.log(`  ${colors.yellow}⚠${colors.reset} ${message}`);
}

module.exports = {
  colors,
  header,
  subheader,
  pass,
  fail,
  info,
  warn
};
