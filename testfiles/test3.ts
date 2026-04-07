import readline from 'node:readline'

// 입출력을 위한 인터페이스 생성
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

// 한 줄씩 입력이 들어올 때마다 이벤트 발생
rl.on('line', (line) => {
  console.log(line + "푸하하");
});