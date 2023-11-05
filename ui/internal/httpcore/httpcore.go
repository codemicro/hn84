package httpcore

import (
	"bytes"
	"git.tdpain.net/codemicro/hn84/ui/internal/config"
	"git.tdpain.net/codemicro/hn84/ui/internal/search"
	"github.com/gofiber/fiber/v2"
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/components"
	"github.com/maragudk/gomponents/html"
	"github.com/uptrace/bun"
	"os"
	"path"
)

type endpoints struct {
	DB *bun.DB
}

func ListenAndServe(db *bun.DB) error {
	app := fiber.New()

	e := &endpoints{db}

	app.Get("/", e.index)
	app.Get("/search", e.search)

	return app.Listen("127.0.0.1:8080")
}

func renderPage(ctx *fiber.Ctx, p g.Node) error {
	b := new(bytes.Buffer)
	if err := p.Render(b); err != nil {
		return err
	}
	ctx.Type("html")
	return ctx.Send(b.Bytes())
}

func (*endpoints) index(ctx *fiber.Ctx) error {
	return renderPage(ctx, basePage("Rummage",
		html.Div(g.Attr("class", "row"),
			html.Div(g.Attr("class", "col-4 mx-auto text-center"),
				html.Img(g.Attr("class", "pb-3"), g.Attr("src", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAJYAAABmCAYAAAApk2j7AAAnuUlEQVR4Ae3BB4DP9eP48efr/Xl/xu3l7DPC2Xvd2dzxMT4IWSEVmcmRKNvZI8kqs6yIxpeMLkSRUZKMZJSdbOfmZ7zfr/99+nbf3+Vvc8ddHg9hjbLxFMRMWxsNtAcCgOvAJWAXsBc4UKF5lV+CC2Z38NQ9EdYoG/9mMdPWOgGVW1CEQJeSNE5bo2z5eequVLKIM/tPlbxxMba1V4DXqgIVCx3iHsS8u84FGEJCQmjbuhWtWrdFURRcLhepVFXFbDYzffq7LFz4Qb69q79/s0LzKhO4g0Ob9oecPXh6BtCcv3kFen8ZkDtwQMnIMof4FxDWKBuZme7Sxdb5m352JjtK83/0QlVDIwqHh27lNmKmrR0MjKVRIZRTN9B/ucR9iAO8AcFtWK1WYmJiWLBgAc2bN+ezzz6jW7du3EQH9gK7StQr/QvwB7A3pEz+M2RywhplI7Ny2Z2Gze/FuPhbr549qBoWztixYzl69ChCcKVq2xpl/XL6nyONmGlrxwKDaVYEQnz5HwEogr9ICULwF02CQcANO3x4gPx5c7F+xTyyBQQgBGz8Zgcder5BqkWLFhEYGEjTpk2pHlaZFas+w81kMhESkhc/b09uxCVgd7qYFD2U5o0i6DNwOD/u28+Va7Hcwm/AHmBnaI3i3wfmDUoEzvnl9L/ME0pYo2xkRlJK5at312menp6MHDmSPbt3svLTz8mbNy8fffQRJpOJ559/nt9//52/LQHsQFdS9akIOvdOETBjD7/v2Yjblu276RI1FDeTyYTT6cTfz4dFi5fiNjZ6BLv3/MS5c+fQdR03Dw8PgoODmTuwHbou6TnlY3Qp+eKjOVQoVw6jqmKxWDAaVVSDyrfffceHyz9nxWdfoOk6uq5zs9LWch65i+dN5gkirFE2MpsdS795Ne5y3LSQkBBD165dkVLyFykZOWoUPt5ebPgyBpfLxbZt2xg2bBi39Fol0CT37FICrPyVNcvm0KJTTzRdx+3zzz7DZDbTpEkTOnXqROvWrXGzWCw0aNCACxcu4HA4SHXt6lXKlS/PZ2O7omk6Zy9dp++7n+BhsTBm6AA6tH4Wh8PJzQyKgkRy+ep1xk15l5WrN5DGxTzF8+YuZS2n8QQQ1igbT4JDm/b76LoeDhQFigDegEKKuIuxatzluHxAacCfFCaTkWHDhqNpGmkpisLkieOJS0hiw4YNWCwWVFWlX//+7PnhBzq2bkaTyNq07/46vFoRJPdGEfD5MTgbi5vJZMThcNK9ezeaNLGhqiqNGzdm8ODBhIWF4aaqKo0bN+bIkSN4enqSSghBmzZt8NGv06VJOG4Gg8JPR84wZtGXKIpC987tiR4ygKSkZG5HVVUuXLjAtPcXsmjFZ7gJIf7ToG+TFjxmwhplI72c2nfCDwgGcgHPXDt31f/CsfMlgepACe7Cz8cLk9GAUAyYDAqXrsdhtztwG/zWm5jMFm5FURQWL/qA47+dYOzYsdSoUQO3cePGERMTg9Go4jKA7FKWeyJ1WPM7nI3Fbc2aNbw/awbrYzYya9YsQkJCMJlMNGzYkIEDB1KjRg1Sde3alaFDBtOiZSvSklKSL18+1kzsQbLDSSrVoPDdgd+Y9vFWNF2nS4fWDB/UD0VRkFJyOx4WM/OXruStURNwy1U0T7Myjcp/wWOicp+u/XG12LVzV/Mf++7X54D6QH7SMBmN5MjmT6CfD/4WM0UK5CJncDayB/mRzc+XQH8f/Lw98fH2xMNs4vK1G1y6ep0tu/ezccdejvx+llSxcQncSmS9OpjMFm5FURSklHR64UVWr17NkCFDyJkjO4sWL2HIkCEcOfIrJ0+egudLc1eajnHhQZwOJwULFuQEsVSpXBm3Bo2asD5mI8l2O25SStySkpJI68UXX+TNQW/SstVzSClJJYSgRLGiTFnxNa+2rEkql6ZTtURBVo0uxPE/LjF+8ToWLFtFUGAAX326hBzZg9F1nZslJdvp8FxzXmzfiq3ffkeXvm+uiZm29kqVNtUqBuQOPEUGU7kPm2Zu6A8MAPwMqgHNpZEiERCAIIXD6RRnzl9Szpy/pADKrn2HBfdIUQQmkxGDwUDZ0iWpXaU81cMrU7p4KN4eZoKLV8Otdt0INE3jZtu2bWPz5s24RUdH06xZM1q0aEF0dDRWq5V/+CUWqgRzW2dvwOpjWHx8iPkyBofDQcOGDcmbJxdu5cqVw01zuXBzuVy4aZpGWhEREUyaNAm73Y7JZCKtmbPfIyIigj6taiIl/+DSdQrkDGLOwPaYVANthi+kYl0bbju/+pR8efKi6To3czpdVA+vyvG93/LTzweDmnV45aSmaVRpXc0rIE9gIhlE5T5EvtpoKjCVO5C6jlAU/kdCwrV4XyBg++KtpYC1pJEtKJB1y97H19cbk9GIrutIKUnL6bCz8D/rcBs0aBCapnGzhMRENm/eTKrhw4fTs0cPcuTMyZAhQ3Db9u1WNn+9lb9UCea24h2w+hjDhg2jSpUqxMfHoygKbna7A7eEhAQURVC6dGmcTidSStq2a0e1atVIy263ExkZSUREPXbs2ImmaaQqWbIkUkrMqkqy08XtOFway4Z3Rpfw9sdbCG/QCpPRyBt9e9LnlRdwOJzczOXSKF2yOKcP7ODQ4aM0aNUpAfjzmcqFaxapXuw46cxQOCyUR0kIwT8IMHmY7CYPU+ylExenNahWtciy6dG0sdXj3J+XOHTkNw4cPkqH55ridDq5lXPnL9Cx10C8PD2pU7cuN1MUhXemTsXpdLLvR5j1Hiz+0Mimzd/j4elJSN68uOXPX4CzZ89w1ZUA5XJwO2LJQV7r2YuaNWuRSkrJmdOnyZkjB6VKl0ZKSY3q1fHx8SVVmdKlMZvN3KxWrVrMnDkTg0FQtWoYqTRNY8eOHSz6YgtNwktyN0JAtVIFaRtREQ+zkemLPmXKzLkYVJXa1arg0jRuJqUkW1AgA/v2oEqFst7z31/22vFdR99KvJ7gylE413bSiaFwWCjp7dwvZ5UfPt5xbvXcKRWaRlTHYFDw8vSkYZ1qdGrVhBkffMz4aXMwm4yEVyqHLiWpFEWhYmQr3EaMGIGu69zsp70/cuDgQfbuhXIVQAiI6qeTL0Th7bePsv/nffj7+3HhwkW2b98O7UuC0cAtqQK+O8voMWNwuVykVbtOHfLkzYvFYsHN18+Pe6FpGiEhIUyaNJkePXpgNBpJVadObd55dwbtIishpeReSKBw3mA6NKhMSK4gZi9bzbh3ZlOiaBGKhRZC1yU303WdkDy5eePVbjSuX0f95utdkTu/2DHyxoXYQNWs7vIK8E7mERLWKBvp6fLJi/X+2HF885aVc0m227kVRVH46PMvmTxnKZqmYVRVWtkacC0+kZhNW3ErGhpK++ef52aqqjJ06FBaNoFP13JLBQrAqVP8V24/aFGY2xLAzB+JiYnBbrfzqCiKQvfu3Tl37hwnT57EYDDgJqUkX758vN2nJflyBCKlJC1FEQgEQvA/upTouiSVIgTX4hIZPGcNl2MT+H7TanLlzIGu69yJn68PbV/uzZebvyFFYvmmlRpmL5RzG4+AsEbZSA+6phs2zlj/p7+fT7btn83D5dK4G6PRSNcB0ezce5CbTZkyhfj4eNJSFIXhw4fjJiV31Lo1fPIJiJy+yObPAAJMCn+R/JcEpITZe+nduxdWa0MeJZPJRFObDafLxYULF3A4HKiqSq5cuUjL19sLu8OO3eHCTQiQkn/o3rwGDasWx6nppOXnaaH18IUYjEYO796Cw+HgbswmE5+uWUffwdE4nS5SrI/o2bCValaTeUDCGmXjUTu89VDk6X0nvqwdXtEwfdTrSCm5V0IITv9xgaYv9iOtPLlz0qNnbzRNw81gMDBx4kTi4+P5cj1YG3FHQnDf5s2bR44cOXiUTCYT7dq2IT4hkd9++w0PDw+yZ8/O/QgtkJujJ/9ACMGq0S+jS/5BAE5No8uEZZQsVoyYT5eSkJjI3SiKwo24OCrVbUpiUhJCEZc9fDw61Xyp3pfcJ2GNsvGoaE6X2DJ34y7NqVUZ2LMzHVs2RNd1HoSqGojZuovXR08jlcVioV7dOtyIi2f37l04nS7MZkhO5o5mz4bevaEO0AqQgAAE/+QAEoCrwPfAd0CO7MF8uGgxTqeTR8VgMPDO1Kls2ryZVLmCA5i7sBv+/t5IKZFSousSIUghEAJ0XZJKUQQR9aJxaRqLh76AxWzkZkIIrt1IoPuk5Vgj6rB0zjskJiVzN4qikJyczKAR4/h0bQxuXoHe63MVy9OsUJUiGvdAWKNsPAoHN/7c/NyhM/8hxefzJ/NMvjxIKXlYJqPKmk3fMXH2Yq5dv87N+vWGqTO5IyEgP9APkNw7FYgGLgHr1q1D0zQeBSEEe/fuZcSIEbjVrl2csWOfx+nUuB+qaqB+5Gik1PloxItouuRWDIrgh1/PMnFpDCWKFmH25GgKFyqIpuncjYfFwoIlKxj99gwSEhJJYS9Q8Zk3itYsMYM7ENYoGw8j7lKsumPZtt1AhcIF8/HFwinYHU4eNSEEHhYzPd4az7e79+FyuUgVswEaNOSW3p4MAwbCDMDF/VOAvsCHH35AYGAQD8tkMvHaq704evx33Dq0rkXXXvV4UA6HC6t1DOO6NaVwSHbuxKQamLL8a7bvP47b26OH0KxRfTw9PdE0jTtRFEFcfAJtXuzJgcNH+dvlSi2qtg3KH/w1NxHWKBsPasvcjaUcifYDQgjeGz+EsPLFyQiqwYDT5aJe255ci40j1apVYLOBxcJfnE4wmSA78BYPRgCDgKGjR1O2bFkelBCCSxcv8nKXLriF5M1NvMPBtYuXiaxZgaGjmyOl5EE0bjSePNl8Gde9GffCpBp4btgCXC6NVNUql+GD2dPx9vJE6joIAUj+R4Lkv7w8PXlj+FgWr/gUp8vF35Lqdm+Qy+RhiiWFsEbZeBAx09b+AeTKkyOIDUtnous6j4NBUTh++iwturxBKi8vCAqCCxfAboeGgJUHI4ABgGIysWbNGhwOB2mZzWY8PDzYvXs3165do1KlSnh4eOB0Okmlqirt27cnNjaWtFq/9iLN2nWgU7X6+Pl5snbtW7hcGvdr3LjP2LhxP5+NfQWXrnM7BkVgMalM/WQ7uw8cIz7Rzp2oBgNCUXATgKIoKIqCyahiMpuwmD2oX6caH3++jviEBNwMRvVsZO+GIcIaZeN+/PTFng4Xf/tzKSmWzhhDmaLPoEvJ46YoCleuxTJg7Ax+/PkQaVUUgo5S8iAEEMV/CaBEyZJ4eXlx+fJlTp06haZp3ExVDTzbvBnJScns3L2bK70mQrHyoGkgJSgK/HkS3mjDjI+X4PDTeb1hZyqUK8Q7776Arkvux/vvbeLjldtZN7knCckO0lINCpeux/He59v5+fg5pJSk5eXpQUJiEm7BBbOPqNC8SjQp7Al2JfbPa8FAbiDvT1/syQ5UB8KAYoDg9hKFNcrGvYqZtvYnoFzDutWYNqIfyXYHTxohBBazidHTFrD08w2k8gE6A6GAxr0RQBT/p2KF8tSrV49q1WuQJ08eNE3DTVEU3JKSkljx0RLen7uAv4SWheHzweXklo4dgLHdmbr8Q5bMmM1PO75n+fK+5MwZwP0YNmwF27f/yoYpvYhLsmMxqvzw6xnW7jjAT0fPkFbXjm3p/lIH8uXNg93hwM1kMjF7wSKiJ00nd/G81Utby+3gPh3fecQXEIXDi8aSQlijbNzN2YOnaxzatH+br5cHq+ZOJke2ADIDD4uZSe8tYsGKL0irG1AK0Lg9AUQBnh4ejBw1iurVq+NwOLgVo9HIyZMn6NmzJ3Fx8bgJQK48AIlx3JYQcPRnGNuD+h2fZePS/1C2bH6mTXuJe+XpaaJq1cG4vd6uHm+v+Jq0/H19qF2zGnOmjkHXJS6Xxq0I4MSp09SytaXmi3XzePp7/cFDENYoG7fjcrjYPPvLFUDbejUqMn3UQFwuF5mNyWhk2RdfM+HduWi6Tqq2QDgg+f+NB0x587J06VIURUHXdYQQSCmRUkcg2LhpE7NnzeDa9RsIAc/Vh5WTYdo8hX4zdVh1EBJucFc+AdChAiQnEejvy6ef9+duhBAgIbL+aFwuF6mMRpWggABWfTCLksWLEp+QyP04/+efVK3fgno9rKrRYtR4QMIaZeN2YqatTQIs+zcuR9d1MjujqrLrp0N0GRBNWuECOkmw819moBf3pmZZWDQUCuYCnPyXCqIm8PkRiL3CPREC9dP3KX3mW6ZO7cytWCxGPD3NRI/8hC+/2kdCQjJuwUEBrJg/g2KhRZCArus8jKPHfyeyRYdYa5TNnwckrFE2bvbT2j0dLx7/c0nNymWYM3EoDqeTrMSoquw9eIQXokYgpSRVOaAboAF9gKY1Yc0E/n86IAGN2wpoBtff3wNOB/dkzxaYOYS1a9/Ey8uCyaSSlOTgo8U7+fnQbxz79Tw3EpJwUwwGxg7uj80aQbZsQTidLh6147+fILJlx5/r92lcjgcgrFE20oqZtnYe0HXy4N40qFMdKSVZlcFg4Jcjx+k97G0uX71GqhzABf5Lbgec3Lcmg2F9QAdo05s7Ugyw9WNYOJU+fRpRsUIh3n57DQcOniat4kWL8Fq3l2jVrCF2uwMpJent9xMnqdOs/UxrlK0P90lYo2ykipm29meDQSmzZ/0SBP8eihDY7XZ6DJnEj/sPk5ZqgIvrIMDMfZm0AgbNV2HhNpCSW1IMMLE3HNpDmTL52b//FGlF1KrGyDf7UaxIYZKSk8loQgjavNyTCzKufZHqxVZwH4Q1yoZbzLS1l42qGnR026fE3ojn38rLy4NX3hjD9h9+xuXSSDX/DejSFNC4Jz9eUKjUTocl34Pm4h+EgNPHEcM7ISX/ExQYQI+XOvFWvx5cj41D13UeN5PJSL5S4eQsmeeZ4nVKneAeCWuUja+mrzsudVno8NersDsdPAWqwcD5y9ep364HqVQDfDIamtcGHNyZBUQ4MGYJ5CsCug5mDzi6D2VcT3R7Mm51aoSxfN4MpJRouo6UkieNw+mkcIVa0hplU7hHIih/cOSVU5c2/rBhGUaFp24ihEBRFDZu203/Ue/gJoBaZeHtV6FiaSCJf1Jg/AoY/B5/EUIgDQq4NKpVrcywAa9SsmgRDKoBTdPJDKInvcv8JSuuW6NsAdwDAVwAsgcF+BGzbAaqwcBTt6YoCgmJSbTsNpDzFy7jZlAUNF0nND/k8ocrcYKDv0tS1akeRuvmjWnWqD6KwYCmaWRGJpORnEUroRiU3fX7NA7jLoQ1ysbP634c8Oex8yMA78lDX6NhnXA0TeepW1NVA6Ui2nEr4ZXL82wTKz1e7kRCQgIul0ZWcfnqVSrUbkLxOiWfzVeu4GruQFijbKS6cOx8/n3rftwJ5Pp+7SJMRpWn/smgKJSu356bfb9pDblyZkdKSVYlhKBd11f5btceyjSqEJiraO5r3IahcFgoqbyDfGILh4W+XaD8MxMH95/S7ciJM17PNqiNS9N4CswmE1WbvYTD6aR5RDirpg/lveVrcfvwo1VE9eyClJKsrH3LZkx7fyF/Hv2jf+Gw0NHchqFwWCg3U1TFVTgsdEqsbv9s6FvTeprMRiqXLYGuS/6tvDwshD/7MiajgQ3zx1KtfAkUAXNXbmDs4P5s/OY7GkbUJltQIFmZruu80LYF73+wzHDu0BnfAhWe+YpbMBQOC+V2PPw8LxaqWmTUjh/2M2783DplShalQN6cSCn5N4mLT6Ry084M7tGeN19pg5QSNyEE81d9SaCfDxs+nkeV+q0Y2Kcbmq6TlXl6erJ91x5OnzobbvIwTfLL6e/iJobCYaHcTVC+bN8UDgsdNWfasgYLPlodUrdaRYIC/JFSkpUpiiDmm110eWMMXy0YS4E8OZBS8j8CFn4SQ7agAJpE1qZj6+ZMe38hNcOrktW93LENk959n8snL75YOCx0KjdRuA91XqlfPV+VZ4JbvjLoUvmGHYmNi0cRgqxINRjo/tYk3l+8iq8/HI+iKNzO9RvxuPn7enP2jwskJSeT1SUkJDJ1zFBS5Nk4c0M9bqJwn/KXf+ayNcqWvVqnWjlrP9edik1ewOlykZX4+nhT1tqBkOz+fDhhAE6Xxq3ousQtMSGBVDPHD6VEWCRGo0pW165VM9x0l7Y5/mq8QhqGwmGhPAijxZRQOCx01KWTF/dMn7P8+U83bKFN0/oYVZXMSgjBvl+OU69tDyYMeJlWDapxNx989hVms5mX2rfETdM0KpcvzdTZC2kYUYesTNcl3To/z6z5izl34FSHQlVDZ/A3hYdUpU21ddYom7Dk8R0a1vRFGnd6jT/+vIQiBJmJwWBgxJQ5vNhvBMsmDySsbDGk5J4kJSWRVnilcuz6cR9Ol4uszsvLk/q1q6HrsvC5Q2cq8jeFR6REvdJjrVE2IYMssxt3jiKyfS/OX7qCIgRPOqOq0qbnYL7YvI1tH71N3lzB3I+kZDtp6brOxlULKFyhFgZFIatbOnc6qqpycOPPe/iboXBYKI9SjkI51xcOCx114sCJ7AuXrq784cq1vNyuKYoQPKnK1G9PfEI8u1dOx+F0cj8Wfb4Rl6bRp2tH0lIUhQB/P/oPHUWXju2QUpJVaZpO2ZKhfL5uI45E++nggjn2KaSTah1r97b2tSlJdvvcCg070eWNaDw9zDxJfLw9KWftgNs3S6YQn5jE/RBCYDQZuZ2OzzUlye7i8/VfkdXVq12LnNmDObP/1CxSKKQngbRG2bp7+Hn22bP/CEVrt+ad+csxGo08TkIIftx/mELVWxCcLZDvVryD3enkfkkpMRmN3MnO9cvpPzgaXdfIyhRF8Mmi90jhQQpD4bBQ0lv+8gW//23X0ZGk+PmXo8xetIrSJYpRIE8OJBlLEYI5Sz9l2JQ5VCwVyqIJr6NpGg9q2RdbcDidvPZKJ25F0zQialenYeuX6NvjJTRNJ6sQQmAyGhk2fiqde73OgiUfE5AnsFaekiGnVTLOoeca1izZuXkk/SfMpffgcWQLCmTle+MI8vdFSkl6Uw0G+g6fwuYde2hUqxKDe7RH0zQehqY5cVMUgaZJbqVY4QL4+XjTudfrzJs2gczOYjazaes2+r41iktXriIEevZCuaaUs1UcxN9UMojRYtx79OS5kn6+XiwY1w9d12nVZwz12vSgVLHCLJ85Bk3TSC9GVaVeu15cuHSFxrUrM7hHO3Rd8rCcmsRNSm5LSskXy96ndK2muDQXqkElszEYDFy5eo1Xot5i9569pNB9s/uts0bZmnELKhkk+zM5fr92PYFUiqLwn9kjuHDlOs/1GU3pyHa82K4Fb3Rrh9Pp4lEym0wUr9sat2cjq/H6y63QdcmjIHWde2E2GWnROJJnytXkwpEfsTscZAZ+vj70fXMkHy7/hL+5ar5YN8DT3yueO1DJODHXb8SPIA0pJdkD/fhu+VS+2vkzo6YvYumqNQzr142W1lpous7DSrY7KN+wI24dmtal5/M2dF3yqGiazr2aMmoQ/9mwmS5Rb/He5GiklDyJPD09WbpiFeOmvcfFS1dIEZunZEj9UvXL/sA9UskgeUqG7D3/6zluRdN1IqqWpuHKaezc+wv9J7zHmHfn8e7IftSoUgFd13kQu/f9QvdBY3Fr07gWPZ+3oeuS9CCl5G40TWPtsvdo8nwPxg95HT8/X54UHhYLm7Z+y9ipszh4+CgprpeMLNOvTNEqiw1Gg859UskgAXmC7LouuROn00Wl0qHs/PgdVm/eRe8hk/H19WLFrPHkzpENKSX3QgjB0s82MPn9Jbh1ahZBt3aN0XVJepFSci9CCxUkd85g2nXpzZefLkVKyeNiNpnYu/8gY6ZM57vdP5LCHlwwx0RrlG0ED0klgymKQNcld+J0aTSuXZmWDarRJ3o2jTq9RniF0syfMgyn08mdGBSFNye9z7qN3+D2SptGdG4Ria5LngRSSn746hPylKmN0ajicDjJSKqqcv7PC/QaOJLdP+zBzWgxncpeOGeh8rZKGo+ISgZThEBHci+S7U4mD3oFi8lIeNt+lI5sxy9fryTZbudWFEXhtWET2bJrH26zR/SmVGhBdF2SHoQQuAnBfUlITKLNs43p0W8w0yeOIr0pioK3lydlajTk9NlzuAkhUBSBrsvkej0aFOARU8lgisEAms79SHY4+fajyfz86wmK1XmOxdNGUrZEKFJKUplNJpp3fZ2jv53Gbd3caHy8PMkIRtXI/Zo0/HUKVIhg7rQJJNvtPGoGgwGXy0nD5zpz+Ohx3MIrlWPFnCkEBvhjNKqMmTqLD5b/ZwnpQCWDGYTgQWiapFSRAqx+bwTNe46kVNFnGDfoVYoXC+WjzzcwYfp8kux2/Hy8WD9vNJqmk94UReBmNBq5Xw6Hk/nvjGXKrLm82rUzj4LJZOLI0WNMmjGHdV9tIUdwNqIH9aF+nWq4uVwaqTRNw8/XnxQhpAOVDCYUwcMI8PVh/bzRNHllGM1e7k9anVvUp1vbxmiaRkZQFAU3i9nIg6hbowq9Bo6kb/eX0DSdB+FhMfP1tp3Mnv8h3+z4gaoVytL75fa8P3E4SXY7bi6Xxq34+vqQIpR0oJLBBA/P29OD3avepd/4uZw+f5FSofkZ3bcziUl2NE0jo1ksFh6Epmns2LCcz77YQPPGVu6FEAKQ7N1/kPZd+pCYlEyRZwrwzug3WTxrIna7A7cku527yebvQwoL6UAlgylC4VFIdjgZ//pLpEpMsvO4eJhNPCg/Hx/eHDWRVs0a43Jp3I7FbGLN+i/pGjUEN5PRyLx3RhNRM4zEpGTc7HYH9yNXrhyk8CIdqGQwoQiyCkUI3ExmMw9KSsmxXTGUDI/k8O4tOBwOUhkMBk6dOk2LF3pw4dJl3Fo3b8TIAb0xm03ouk5iUjIPKmdwNlJYSAcqTz0wIQRuXh4WHkZ8YiIB/v6UCIugT7cX8fPz45tt21n31VZcmobFbGbxrInUCquEw+nETdd1Hlb24CAEmEkHKhnMYDCQVQj+y8PDwsNat3wuxcMbMmriNFI1axRB9MBX8fbyRNclDqeTRylbgD+S9KGSwZxOF1mFogjcDAYDD8sg4MsV8xg/fS4vtW9J3RpVsdsduOm6JD1IKUkvKhnsyvVYzCYTWUneXDl5FJ4pEMK8qaNxs9sdpDddl6QXhYzl+OPiNbKa3LlykRnpukZ6UchYl67FxZFVCKHglj9vDjIjTddJLwoZ63zsjUSympBcwTz1TwoZ6/jVG3FkNTly5iIz0nSd9KKQgQxGw6/7Dp8gq8mdIxtP/ZNCBgqtUfz3g0dPkNV4mM1kRk6nk/SikIGCC2RfceVaLMkOJ1mJrutkRrouSS8KGcjDz9OZp2RIhQYvvcXOnw6jKIKsQEpJZiSlJL0oZLBS9cv+ZO3bRAyZvmhvm75jSUhMRghBZiSlTmZnMhlJDwqPgxBE9GpY0ZwvoFKjV4Y6nusTjaIoPJXxNF2SHhQeo9AaxX+MfLWR5cKV2FE12vdnwryVGI0qT2UMRVHQXC7Sg8JjZlAN0hplG1nx2Sre67d+f752hwHs+/U3hBA86XRdkplZzCZSJJAOFJ4Q2QpkT7BG2XLnq1TI9tro92jffzyqQUEInli6lLgpikJmtHn792TLHzybdKDwhClUtcg6a5RNJJn4uFq7/sxatg7VoPAk0qXETQhBZqIoCsd+P03n3gMvVmxRdSDpQOUJVallWDug3cqZ6w+tWLelxOyRr1E6ND9SSp4Uuq7j5tI0MgMhBJpLo1i1Rpi8LNusUbZapBOVJ1z9VxuXdCQ5fHuNnB5rMZv4ZslkHC4nUvLYaZqO27k/L+HtaeFJpaoGLl25RkTLF4lPSDxWr6e1rNFsTCIdqWQCJg/TDWuUTZz/9VzDqm36bihSIA+LJw7A6dJ43BRFcP7cOYoUKcSTRAiBajTStd9QNn+zgxQba3eJaGvx8bhGBhDWKBuZzfbFW6MTrsYPax4RztCe7UmyO3hcWvcdQ/SgvtStGc7jJoTA08PCrIUfMWPBMuLi4jF7W+bV6RrZjQymkgnVeKHOcGD46mlrN63evDPize5taV4vDKdLI6MF+fty+tx5HieT0chPB36hx6BoLly4RIrknEVyRVfrUmc8j4lKJmaNskU6k52miXNX/jlhzscB88f0o3ihEDRdJ6MEBfhz+PgJMpoiBE6XRvcBI/l25/e4Karyh1egd/4aL9Rx8ZipZHJGi9HR4LUmgfu+2BPedeg7O0xGlc2LJuAmJeku0M+bK1evk1HMJiM/7DtEx15vkJxs52+zIno27K+aVQdPCGGNspFVJMcliW8WbB4LvJU9yJ9FE17H28sTKSXpZcnqzfz462lWzH2b9CKEwMfLk+d7vMGW73YjpSTFlcqtwhsHhgR9zxNIJQux+HhIa5RtsNT1YV9NX3+40SvDihQpkJfpQ3vi7WlBSsmj5uvtyfUbcaQH1WDgl6PHGTByMr8e+x233CVCRuYtGTIuIE+gkyeYShYkFEWzRtlCz+w/VfWXrw/ENOo6xC93cAALJwzAy8OMlDwyZpORa9djeZSMRpWvt+2ixxujcDqdmL0tW/OVyf9S8XqlT5JJqGRhIWXy7w4pk99/++KtHf+4dG1Jwy5DKJAnOx9MGICqKEgenp+3B7GxN3gUBIL5H33GxOlzSBHnFeg9rN4L1nfJhFT+BWq8UGep5tKWbpq5IebkuYsN6nYaSHCgHzELxxGfkMTD8PLyRpcSVTXgcmk8CH8/XypGPseZc3+Q4mjkq43KGlRDMpmYyr+EQTVgjbJZE68neG/7cMvaS1dja1d4tjdFCuRh0YTXcWk6D8LH04KbyWjC5UriXimKwqXLV2nX/XVOnTlH8TqlnqvdqMR/LN4WjSxA5V/G098r3hplq3Nq34mgX7ce+ubYyXMlq7XrT4ki+Xnnze54epiRUnKvfL08cPOwmElMSuJujEaVdZu+ZcjYacTeuPFTsTolu1hblf+JLEblXyp/uYJX8pcrWOq33ce8ju88suWXY6cqW7sMJjDAj3nRr5EjyB9dSu7Gz9eLe2E0qixdtYZhE6aTrUD2VUUiS3QMypfNQRal8i9XqGqRhEJVi1S5dOKi4ci3hw5cvRZbvFWf0RhUA6tnjcDfxwtdSm7H29MDt03bd1O+ZFHSEkKg6zrDJs1k5efrNJ9sPiusUbaO/AuoPPWX4ILZteCC2UtIXSpfTV/3rubSXrV1H46iKGz6YDxGo4qUktvZ+PW3lC9ZlFReXp7kL18Pl8tFitEN+jYZIYSQ/EsIa5SNp27txA/Hyx397tcdgIe3lwdTBnWjdJH8aLpOKk8PM5Va9iFXrhxsW72EuPhEmr/Qi9Nn/0AoysiIntYxBqNB419GWKNsPHV33y355pX4K3FzPSxmovt2JrxcUXRdEp+YRONXhmEyGfH18ebylWuk2GKNstXjX8xQOCyUp+4uX9kCewuHhY76dfvhVhu/+zHHtr2HaREZjrenB/NXfYmm6SQmJZO3VL6I8OdrjuRfTlijbDx1f3Ys/faluMs35gMKf/PN7rcr/Pma4Tz1l/8HlVYS0KLshlUAAAAASUVORK5CYII="), g.Attr("alt", "Clipart box of stuff")),
			),
		),
		searchBox(""),
	))
}

func (e *endpoints) search(ctx *fiber.Ctx) error {
	queryStr := ctx.Query("q")
	if queryStr == "" {
		return ctx.Redirect("/")
	}

	queryTokens := search.PlaintextToTokens(queryStr)
	results, err := search.DoSearch(e.DB, queryTokens)
	if err != nil {
		return err
	}

	var resultNodes []g.Node

	conf := config.Get()

	for i, res := range results {
		plaintext, err := os.ReadFile(path.Join(conf.CrawlDataDir, res.Document.ID+".txt"))

		var text g.Node

		if err == nil {
			middleToken := res.Tokens[len(res.Tokens)/2]
			startPos := middleToken.Start - 50
			endPos := 100 - (middleToken.End - middleToken.Start) + middleToken.End
			if endPos >= len(plaintext) {
				endPos = len(plaintext) - 1
			}

			x := make([]g.Node, endPos-startPos)

			for _, tok := range res.Tokens {
				if tok.Start >= startPos && tok.End <= endPos {
					for i, ch := range plaintext[tok.Start : tok.End+1] {
						x[(tok.Start-startPos)+i] = html.B(g.Text(string(ch)))
					}
				}
			}

			for i, b := range x {
				if b == nil {
					x[i] = g.Text(string(plaintext[startPos+i]))
				}
			}

			text = g.Group(x)
		}

		class := "mt-4 p-1 border-radius-1"
		if i%2 == 0 {
			class += " bg-primary-subtle"
		}

		resultNodes = append(resultNodes, html.Div(
			g.Attr("class", class),
			html.H4(unstyledLink(res.Document.Title, res.Document.URL)),
			html.Span(g.Attr("class", "text-success"), html.A(g.Text(res.Document.URL), g.Attr("href", res.Document.URL), g.Attr("class", "text-decoration-none text-success"))),
			html.Br(),
			html.Span(g.Attr("class", "text-secondary fst-italic text-sm"), html.A(text, g.Attr("href", res.Document.URL), g.Attr("class", "text-decoration-none text-secondary"))),
		))
	}

	return renderPage(ctx, basePage("Rummage search results",
		searchBox(queryStr),
		g.Group(resultNodes),
	))
}

func unstyledLink(text string, href string) g.Node {
	return html.A(g.Text(text), g.Attr("href", href), g.Attr("class", "text-decoration-none"))
}

func basePage(title string, content ...g.Node) g.Node {
	return components.HTML5(components.HTML5Props{
		Title:    title,
		Language: "en-GB",
		Head: []g.Node{
			html.Link(g.Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css"), g.Attr("rel", "stylesheet")),
		},
		Body: []g.Node{
			html.Nav(g.Attr("class", "navbar bg-dark border-bottom border-body"), g.Attr("data-bs-theme", "dark"),
				html.Div(g.Attr("class", "container-fluid"), html.A(g.Attr("class", "navbar-brand"), g.Attr("href", "/"), g.Text("Rummage"))),
			),
			html.Div(g.Attr("class", "pt-4 container"), g.Group(content)),
			html.Script(g.Attr("defer", "true"), g.Attr("src", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js"))},
	})
}

func searchBox(query string) g.Node {
	return html.FormEl(
		g.Attr("action", "/search"),
		html.Div(
			g.Attr("class", "input-group input-group-lg"),
			html.Input(
				g.Attr("type", "text"),
				g.Attr("class", "form-control"),
				g.Attr("placeholder", "Enter your query..."),
				g.Attr("name", "q"),
				g.If(query != "", g.Attr("value", query)),
			),
			html.Button(
				g.Attr("class", "btn btn-outline-primary"),
				g.Attr("type", "submit"),
				g.Text("Submit"),
			),
		),
	)
}
