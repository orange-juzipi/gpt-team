type RandomSource = Pick<Crypto, "getRandomValues">

const ACCOUNT_ALPHABET = "abcdefghijklmnopqrstuvwxyz"
const PASSWORD_LOWERCASE = "abcdefghijklmnopqrstuvwxyz"
const PASSWORD_UPPERCASE = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const PASSWORD_DIGITS = "0123456789"
const PASSWORD_SYMBOLS = "!@#$%^&*-_=+?"
const PASSWORD_CHARSETS = [
  PASSWORD_LOWERCASE,
  PASSWORD_UPPERCASE,
  PASSWORD_DIGITS,
  PASSWORD_SYMBOLS,
]
const PASSWORD_ALPHABET = PASSWORD_CHARSETS.join("")

function resolveRandomSource(randomSource?: RandomSource) {
  if (randomSource) {
    return randomSource
  }

  if (!globalThis.crypto) {
    throw new Error("当前环境不支持安全随机密码生成。")
  }

  return globalThis.crypto
}

function randomIndex(limit: number, randomSource: RandomSource) {
  if (limit < 1) {
    throw new Error("limit must be greater than 0")
  }

  const maxValidByte = 256 - (256 % limit)
  const bytes = new Uint8Array(1)

  while (true) {
    randomSource.getRandomValues(bytes)
    if (bytes[0] < maxValidByte) {
      return bytes[0] % limit
    }
  }
}

function randomCharacter(charset: string, randomSource: RandomSource) {
  return charset[randomIndex(charset.length, randomSource)]
}

export function generateRandomAccount(domainSuffix: string, random = Math.random) {
  const length = Math.floor(random() * 4) + 5

  let prefix = ""
  for (let index = 0; index < length; index += 1) {
    prefix += ACCOUNT_ALPHABET[Math.floor(random() * ACCOUNT_ALPHABET.length)]
  }

  return `${prefix}@${domainSuffix}`
}

export function generateSecurePassword(length = 18, randomSource?: RandomSource) {
  if (length < PASSWORD_CHARSETS.length) {
    throw new Error(`密码长度不能小于 ${PASSWORD_CHARSETS.length} 位。`)
  }

  const source = resolveRandomSource(randomSource)
  const characters = PASSWORD_CHARSETS.map((charset) => randomCharacter(charset, source))

  while (characters.length < length) {
    characters.push(randomCharacter(PASSWORD_ALPHABET, source))
  }

  for (let index = characters.length - 1; index > 0; index -= 1) {
    const swapIndex = randomIndex(index + 1, source)
    ;[characters[index], characters[swapIndex]] = [characters[swapIndex], characters[index]]
  }

  return characters.join("")
}
