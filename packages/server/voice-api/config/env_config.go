package config

type Config struct {
	DB struct {
		Host     string `env:"DB_HOST,required"`
		Port     int    `env:"DB_PORT,required"`
		User     string `env:"DB_USER,required,unset"`
		Name     string `env:"DB_NAME,required"`
		Password string `env:"DB_PASSWORD,required,unset"`
	}

	AWS struct {
		Region string `env:"AWS_S3_REGION,required"`
		Bucket string `env:"AWS_S3_BUCKET,required"`
	}

	Service struct {
		ServerAddress string `env:"VOICE_API_SERVER_ADDRESS,required"`
		ApiKey        string `env:"VOICE_API_KEY,required"`
		Host          string `env:"VOICE_API_HOST,required"`
	}
}
