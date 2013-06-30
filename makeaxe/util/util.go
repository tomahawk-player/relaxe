/* === This file is part of Tomahawk Player - <http://tomahawk-player.org> ===
 *
 *   Copyright 2013, Teo Mrnjavac <teo@kde.org>
 *
 *   Tomahawk is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   Tomahawk is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with Tomahawk. If not, see <http://www.gnu.org/licenses/>.
 */

package util

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
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
	if contents, err := ioutil.ReadFile(filePath); err == nil {
		h := md5.New()
		h.Write(contents)
		sum := h.Sum(nil)
		output := fmt.Sprintf("%x", sum)
		return output, nil
	}
	return "", fmt.Errorf("Cannot open file %v to compute MD5 sum.", filePath)
}
