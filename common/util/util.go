/* === This file is part of Relaxe - <https://github.com/teo/relaxe> ===
 *
 *   Copyright 2013, Teo Mrnjavac <teo@kde.org>
 *
 *   Relaxe is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   Relaxe is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with Relaxe. If not, see <http://www.gnu.org/licenses/>.
 */

package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ExistsDir(path string) (bool, error) {
	ex, err := exists(path)
	if !ex || err != nil {
		return ex, err
	}

	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return st.IsDir(), nil
}

func ExistsFile(path string) (bool, error) {
	ex, err := exists(path)
	if !ex || err != nil {
		return ex, err
	}

	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !st.IsDir(), nil
}

func Md5sum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err == nil {
		defer f.Close()

		h := md5.New()
		if _, err := io.Copy(h, f); err == nil {
			sum := h.Sum(nil)
			return fmt.Sprintf("%x", sum), nil
		}
	}
	return "", fmt.Errorf("Cannot open file %v to compute MD5 sum.", filePath)
}

func maxInt(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

// returns -1 if first is less than second, 1 if first
// is more than second, and 0 if they are equal
func VersionCompare(first string, second string) (verdict int) {
	verdict = 0

	if first == second {
		return
	}

	sFirst := strings.Split(first, ".")
	sSecond := strings.Split(second, ".")

	depth := maxInt(len(sFirst), len(sSecond))

	if lf := len(sFirst); lf < len(sSecond) {
		for i := 0; i < depth-lf; i++ {
			sFirst = append(sFirst, "0")
		}
	}

	if ls := len(sSecond); ls < len(sFirst) {
		for i := 0; i < depth-ls; i++ {
			sSecond = append(sSecond, "0")
		}
	}

	for i := 0; i < depth; i++ {
		a, er1 := strconv.ParseUint(sFirst[i], 10, 16)
		b, er2 := strconv.ParseUint(sSecond[i], 10, 16)

		if er1 == nil && er2 == nil {
			if a < b {
				verdict = -1
				break
			} else if b < a {
				verdict = 1
				break
			}
		} else { //fallback: string comparison
			if sFirst[i] < sSecond[i] {
				verdict = -1
				break
			} else if sSecond[i] < sFirst[i] {
				verdict = 1
				break
			}
		}
	}
	return
}
