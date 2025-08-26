module.exports = {
  default: [
    'features/**/*.feature',
    '--require features/step_definitions/**/*.js',
    '--require features/support/**/*.js',
    '--format progress-bar',
    '--format json:cucumber-report.json'
  ].join(' ')
};