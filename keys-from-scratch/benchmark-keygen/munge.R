#!/usr/bin/env Rscript
library(tidyverse)
ciphers <- c("3DES", "CAST5", "AES128", "AES192", "AES256")
compressionAlgos <- c("None", "ZIP", "ZLIB")
levels <- c("NoCompression", "BestSpeed", "BestCompression", "DefaultCompression")
rsaBits <- c(2048, 4096)
hashes <- c("SHA224", "SHA256", "SHA384", "SHA512")
args <- commandArgs(trailingOnly = TRUE)
test <- read_csv(
  args[1],
  col_names = c(
    "Duration", "Hash", "Cipher", "Compr", "Level",
    "RSA"
  ),
  col_types = cols_only(
    Duration = col_number(),
    Hash = col_factor(
      levels = hashes,
      ordered = FALSE,
      include_na = FALSE
    ),
    Cipher = col_factor(
      levels = ciphers,
      ordered = FALSE,
      include_na = FALSE
    ), Compr = col_factor(
      levels = compressionAlgos,
      ordered = FALSE,
      include_na = FALSE
    ),
    Level = col_factor(
      levels = levels,
      ordered = FALSE,
      include_na = FALSE
    ),
    RSA = col_factor(
      levels = rsaBits,
      ordered = FALSE,
      include_na = FALSE
    )
  )
)
testDf <- as_tibble(test)
testDf <- arrange(testDf, Hash, Cipher, Compr, RSA, Level) %>%
  mutate(
    Duration = Duration / 10^9,
    Level = factor(gsub("ession|ee|ault", "", Level))
  ) %>%
  select(
    RSA, Cipher, Hash, Compr,
    Level, Duration
  )

noComprWithLevel <- filter(testDf, as.character(Compr) == "None") %>%
  group_by(
    Hash,
    Cipher,
    RSA,
    Level
  ) %>%
  summarize(
    medianDur = median(Duration),
    meanDur = mean(Duration),
    durationSd = sd(Duration),
    coeffOfVar = (durationSd / meanDur) * 100
  )


noComprWithWoLevel <- filter(testDf, as.character(Compr) == "None") %>%
  group_by(
    Hash,
    Cipher,
    RSA
  ) %>%
  summarize(
    medianDur = median(Duration),
    meanDur = mean(Duration),
    durationSd = sd(Duration),
    coeffOfVar = (durationSd / meanDur) * 100
  )


summariseLevel <- group_by(testDf, Hash, Cipher, Compr, RSA) %>%
  summarize(
    medianDur = median(Duration),
    meanDur = mean(Duration),
    durationSd = sd(Duration),
    coeffOfVar = (durationSd / meanDur) * 100
  )

summariseLevelCompr <- group_by(testDf, Hash, Cipher, RSA) %>%
  summarize(
    medianDur = median(Duration),
    meanDur = mean(Duration),
    durationSd = sd(Duration),
    coeffOfVar = (durationSd / meanDur) * 100
  )

everything <- group_by(testDf, Hash, Cipher, Compr, Level, RSA) %>%
  summarize(
    medianDur = median(Duration),
    meanDur = mean(Duration),
    durationSd = sd(Duration),
    coeffOfVar = (durationSd / meanDur) * 100
  )

fresh <- tibble(
  name = c(
    "ComprNoneWLvl",
    "ComprNoneWoLvl",
    "SumLvl",
    "SumLvlCompr",
    "All"
  ),
  min = sapply(
    c(
      noComprWithLevel[, "coeffOfVar"],
      noComprWithWoLevel[, "coeffOfVar"],
      summariseLevel[, "coeffOfVar"],
      summariseLevelCompr[, "coeffOfVar"],
      everything[, "coeffOfVar"]
    ),
    min
  ),
  max = sapply(
    c(
      noComprWithLevel[, "coeffOfVar"],
      noComprWithWoLevel[, "coeffOfVar"],
      summariseLevel[, "coeffOfVar"],
      summariseLevelCompr[, "coeffOfVar"],
      everything[, "coeffOfVar"]
    ),
    max
  ),
  mean = sapply(
    c(
      noComprWithLevel[, "coeffOfVar"],
      noComprWithWoLevel[, "coeffOfVar"],
      summariseLevel[, "coeffOfVar"],
      summariseLevelCompr[, "coeffOfVar"],
      everything[, "coeffOfVar"]
    ),
    mean
  ),
  median = sapply(
    c(
      noComprWithLevel[, "coeffOfVar"],
      noComprWithWoLevel[, "coeffOfVar"],
      summariseLevel[, "coeffOfVar"],
      summariseLevelCompr[, "coeffOfVar"],
      everything[, "coeffOfVar"]
    ),
    median
  )
)

print.data.frame(fresh)

fasterKeys <- filter(everything, coeffOfVar < 50, medianDur < 5) %>%
  arrange(coeffOfVar, medianDur, desc(RSA)) %>%
  mutate_if(is.numeric, signif, digits = 4)

print(nrow(everything))
print(nrow(fasterKeys))
print.data.frame(fasterKeys[1:20, ])

fastHighRsa <- filter(
  everything,
  as.character(RSA) == "4096",
  medianDur < 5,
  as.character(Compr) != "None",
  as.character(Level) != "NoCompr",
  coeffOfVar < 40
) %>%
  arrange(medianDur) %>%
  mutate_if(is.numeric, signif, digits = 4)

print.data.frame(fastHighRsa)

split <- arrange(everything, Hash, Cipher, Compr, Level, RSA)
smallBits <- pull(filter(split, as.character(RSA) == "2048"), medianDur)
bigBits <- pull(filter(split, as.character(RSA) == "4096"), medianDur)

bitPercentIncrease <- (100 * (bigBits - smallBits) / smallBits)

print.data.frame(tibble(
  name = c("medianDur % inc"),
  min = c(min(bitPercentIncrease)),
  max = c(max(bitPercentIncrease)),
  mean = c(mean(bitPercentIncrease)),
  median = c(median(bitPercentIncrease))
))
